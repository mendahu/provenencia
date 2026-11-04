# Ship log — S13-08 Unpromoting the last member

**Plan:** [`S13-08-unpromoting-the-last-member.md`](S13-08-unpromoting-the-last-member.md) · **Branch:** `s13-08-unpromoting-the-last-member` · **PR:** [#707](https://github.com/mendahu/provenencia/pull/707)
**State:** ready

## Round 0 — implement

Head `637e81b9f7550c60df5fde09a34946f44730ed4b`. Gate PASS (orchestrator re-ran `scripts/check-pr.sh`). Report: none.

## Round 1 — review

Reviewers: CON (Composer), MNT (Grok 4.7), UX (Grok 4.7), SEC+PRF (Grok 4.7).
Skipped: none — UX matches `macos/App/**` / `Localizable.xcstrings`; SEC+PRF matches `api/ffi/**`, `core/database/**`, Features models.

Base `62b94f03c45ebfefb9542f03d0eeb50a365b54af` · Head `637e81b9f7550c60df5fde09a34946f44730ed4b`

| ID | Lens | Severity | Title | Decision | Evidence / reason |
| --- | --- | --- | --- | --- | --- |
| CON-1 | CON | major | `TestGetUnpromoteImpact` omits bridge and merged-into | Fix now → F-3 | verified |
| CON-2 | CON | major | AC-1.5 merged-into has no Go proof | Fix now → F-1 | verified |
| CON-3 | CON | major | Subject delete bridge sole-member missing | Fix now → F-4 | verified |
| CON-4 | CON | major | AC-1.1 write effects (`Handles` / `counts`) not asserted | Fix now → F-5 | verified |
| CON-5 | CON | major | T-6 only exercises one fixture | Fix now → F-6 | verified; later CON-7 for AC-1.7 row |
| CON-6 | CON | major | T-2 Event/Place sole-member omitted | Fix now → F-2 | verified |
| MNT-1 | MNT | major | §5.2 still says a departing claim touches no other claim | Fix now → F-7 | verified |
| MNT-2 | MNT | minor | `ViaDeletesCitations` unused by facet | Fix now → F-8 | verified |
| UX-1 | UX | major | Return deletes person while focus is on Keep | Fix now → F-11 | user 2026-11-04: fix all three |
| UX-2 | UX | major | “Handle stays” while association deleted | Fix now → F-12 | user 2026-11-04 |
| UX-3 | UX | major | Failed preflight still says handle stays | Fix now → F-13 | user 2026-11-04 |
| UX-4 | UX | minor | Second Unpromote reopens after Cancel | Fix now → F-9 | verified |
| SEC-1 | SEC | major | Subject cascade loads every claim before list cap | Fix now → F-10 | verified |

Nits (logged): `CON · plan §9 DoD checkboxes still [ ]` · `CON · T-1 seven fields on canonical_entity delete change not asserted` · `MNT-3 · bridgefile.go:95 · “Does not commit.” on one line` · `MNT-4 · fts.go:498 · use Conclusion's ok result`

### Fix plan 1

F-1–F-13 as previously recorded (tests, docs, danger tone, full UnpromoteImpact, preflight-failed copy, SQL LIMIT, generation guard, ViaDeletesCitations).

## Round 1 — fix

Head `d68e4ad5d0b5d9028da9dcdcfd36421a6b917760`. Gate PASS (orchestrator). F-1–F-13 done (F-6 later completed in F-2.3). Deviations: Landed note records D-11 / Out / AC-2.6 / D-15 amendments.

## Round 2 — review

Reviewers with Fix-now blocker/major: CON, MNT, UX, SEC+PRF.
Scope: `637e81b9…d68e4ad5`. Round 2 may only add blockers.

Round-1 resolutions: CON-1–4, CON-6 **Resolved**; CON-5 **Not resolved** (→ CON-7). MNT-1 **Resolved**. SEC-1 **Resolved**. UX-1–4 **Resolved**. MNT-2 resolved in F-8.

| ID | Lens | Severity | Title | Decision | Evidence / reason |
| --- | --- | --- | --- | --- | --- |
| CON-7 | CON | major | T-6 still omits AC-1.7 fixture | Fix now → F-2.3 | incomplete F-6 |
| MNT-5 | MNT | blocker | Unpromote preflight builds group beside `ListedGroup` | Fix now → F-2.1 | verified `GroupFromProbeRows` / `DeletedHandleRef` |
| MNT-6 | MNT | blocker | Unpromote confirm copies the cascade sentence | Fix now → F-2.2 | verified `unpromoteDeletedHandlesMessage` |

### Fix plan 2

F-2.1 ListedGroup assembly; F-2.2 cascadeLines; F-2.3 T-6 AC-1.7 row.

## Round 2 — fix

Head `9e2d4aef6ae8a88ad4358eab7524951a831dd256`. Gate PASS (implementer). F-2.1–F-2.3 done.
Orchestrator verified: `UnpromoteImpact` → `ListedGroup`; confirm uses `PVDeleteImpactCopy.cascadeLines`; T-6 has provisional-only row; `GroupFromProbeRows` / `DeletedHandleRef` / `unpromoteDeletedHandlesMessage` gone.
Merged `origin/main` → head `bbbd71a7…`. Gate PASS against `82464c6a`.

## Result

Gate PASS · CI green (Go suite flaked once on `TestLargeSourceBudgets/warm` 355ms/350ms; rerun pass; macOS pass) · open: none · follow-ups: nits only (DoD checkboxes, seven-field assert, comment reflow, Conclusion ok) · to user: — · CON miss: none
