# Ship log — S10-61 Cap region size

**Plan:** [`S10-61-region-caps.md`](S10-61-region-caps.md) · **Branch:** `s10-61-region-caps` · **PR:** #462
**State:** ready

Base: `main` @ `721245e9`. Worktree: `/Users/jakerobins/Repos/provenencia-s10-61`.

## Round 0 — implement
Head `fee0cf74`. Gate PASS (orchestrator rerun). Report: no deviations; `RegionFixtures.swift` added to `project.pbxproj` by hand; manual check (§10) not run. PR #462 (draft).

## Round 1 — review
Verdicts: CON changes needed · MNT clean · UX clean · SEC clean · PRF clean.

| ID | Lens | Severity | Title | Decision | Evidence / reason |
| --- | --- | --- | --- | --- | --- |
| CON-1 | CON | major | Plan has no `## Landed` note though the code differs from the plan | Fix now → F-1 | verified: plan headings end at §13; implement-pr step 8 requires the note where code differs; differences exist (`paddedArtifactDoc`, T-5 rows via `doc(...)`, extra D-5 sentence in §3.3) |
| MNT-2 | MNT | nit | `paddedArtifactDoc` not named in Codebase fit | Fix now → F-1 (folded) | recorded in the Landed note, no code change |
| MNT-1, CON-2, UX (outside lens) | MNT | minor | §3.3 bullet says "the app writes one region per locator" twice | Fix now → F-2 | verified: `docs/interpretation-layer-data-model.md:284`; F-1 is already a docs commit, and this is a one-line merge |
| UX-1 | UX | minor | Point cap's only signal is a beep with no reason (no VoiceOver/status) | Follow-up | D-8 rejects new chrome for this PR on purpose; idea: accessibility announcement or status line via L10n |

### Fix plan 1
**F-1 — Add the plan's Landed note** (from CON-1, MNT-2)
Files: `docs/deployment-plan/spike-10-followups/prs/S10-61-region-caps.md`.
Failing test first: none (docs only).
Do: add a `## Landed` section after the last section, in the shape of `docs/deployment-plan/spike-11/prs/S11-06-migrate-reconciler.md` ("Differences from the plan:"). List: the test-only helper `paddedArtifactDoc` in `core/locator/validate_test.go` (used by the T-3 byte-cap rows; not named in D-11); the T-5 rows built with the existing `doc(...)` builder rather than `fmt.Sprintf`; `RegionFixtures.swift` wired into `project.pbxproj` by hand; the §3.3 wording after F-2. Say the §10 manual check is left for the user.
Check: `python3 scripts/check-plan-docs.py` 0 errors; scripts/check-pr.sh PASS.

**F-2 — Say the one-region-per-locator clause once in §3.3** (from MNT-1, CON-2)
Files: `docs/interpretation-layer-data-model.md:284`.
Do: merge the bullet's two sentences so it reads, in substance: "a locator document holds at most 256 region points in total, counted across its region selectors (`locator.MaxRegionPoints`), and is at most 64 KiB (`locator.MaxDocumentBytes`). The app's freeform tool and polygon check cap each polygon at the same number, which equals Go's per-document count because the app writes one region per locator." Leave the Swift doc comment as is.
Check: scripts/check-pr.sh PASS.

## Round 1 — fix
Head `0db3aa06`. Gate PASS (orchestrator rerun). F-1 done (`0db3aa06`); F-2 done (`8cb180c4`).

## Round 2 — review
Lenses: CON, MNT (the lenses with Fix-now findings). Scope `fee0cf74...0db3aa06`.
Verdicts: CON clean · MNT clean. CON-1, CON-2, MNT-1, MNT-2 resolved; no new findings.

## Result
Gate PASS · CI see PR #462 · open: none · follow-ups: UX-1 · to user: —
