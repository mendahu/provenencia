# Ship log — S12-20 Docs and skills match the code

**Plan:** [`S12-20-docs-match-code.md`](S12-20-docs-match-code.md) · **Branch:** `s12-20-docs-match-code` · **PR:** #591
**State:** ready

## Round 0 — implement
Head `4037c23b`. Gate PASS (orchestrator run, against `5312337a`). Report: no deviations. Base moved `67988a5` → `5312337a` (S11-28); the S11-28 diff touched `conclusion-reconciliation.md`, `matching-model.md`, `seeded-vocabulary.md` but none of the lines this plan edits.

## Round 1 — review
Reviewers: CON (Sonnet), MNT, UX, SEC+PRF.
- UX: diff touches `macos/App/DesignSystem/README.md` (path under `macos/App/**`).
- SEC+PRF: diff touches `core/database/effects/README.md` (path under `core/database/**`) and `scripts/check-guidance-names.py`; docs only, dispatched because the path matches.

All four reviewers: Clean. SEC+PRF and UX: no findings.

| ID | Lens | Severity | Title | Decision | Evidence / reason |
| --- | --- | --- | --- | --- | --- |
| CON-1 | CON | minor | Plan has no Landed note | Fix now → F-1 | verified: Status line 3 is Landed, no `**Landed:**` line; S12-15 plan line 4 shows the convention; DoD line 324 asks for it |
| MNT-1 | MNT | minor | `catalog-deletes.md` gate enum is still a hand copy | Fix now → F-2 (pointer only) | verified: line 73 lists the `Gate` constants (`core/database/deleteimpact/kinds.go`) with no owner named. D-5(a) keeps the enum, so the fix adds the owner to the comment instead of removing the list. Line 11's list is a line this PR does not touch, and that line already names `extraGateList`: not taken |
| MNT-2 | MNT | minor | Script docstring summary misses the kit check; "It cannot see" now reads as the kit check | Fix now → F-3 | verified: `scripts/check-guidance-names.py` lines 2 and 17–22 |
| MNT-3 | MNT | minor | `matching-model.md` line 3 still says "kept true as the layers land" | Fix now → F-4 | verified: line 3 at head; D-4 removes the not-built wording |

Nits (logged): `CON-2 · .cursor/rules/catalog-deletes.mdc:12 · sentence starts lowercase "infra"` · `MNT-4 · scripts/check-pr.sh:19 · wrap the widened guidance-names comment to ~80 columns` · `MNT-5 · scripts/check-guidance-names.py:38 · move KIT/KIT_LAYERS out of the regex block`

### Fix plan 1
No blockers or majors; all four are minors in files this PR changes. Docs only, so there are no failing tests; each Check is a grep.

**F-1 — Add the plan's Landed note** (from CON-1)
Files: `docs/deployment-plan/spike-12/prs/S12-20-docs-match-code.md`.
Do: under the Status line, add a `**Landed:**` line in the shape of S12-15's (`S12-15-graph-search.md` line 4): "Differences from the plan: the base moved from `67988a5` to `5312337a` (S11-28), which touched none of the lines this plan edits. Review round 1 added: an owner pointer on the `catalog-deletes.md` gate enum, a widened `check-guidance-names.py` docstring, and the dropped "as the layers land" clause in `matching-model.md`. The manual check (section 10) was not run in the app by the implementer."
Check: `rg -n '^\*\*Landed:\*\*' <plan>` → one hit; `python3 scripts/check-plan-docs.py <plan>` → 0 errors.

**F-2 — Name the owner of the gate enum** (from MNT-1)
Files: `docs/catalog-deletes.md` line 73.
Do: keep the enum as D-5(a) wrote it; append to the comment ` (the `Gate` constants in `core/database/deleteimpact/kinds.go`)`, or put that on its own comment line just below if the line runs too long. Do not change line 11.
Check: `rg -n 'deleteimpact/kinds.go' docs/catalog-deletes.md` → one hit; the AC-2.4 grep `rg -n 'edge_locked \| last_graph \| infra' docs/catalog-deletes.md` still → one hit.

**F-3 — Make the script docstring cover both checks** (from MNT-2)
Files: `scripts/check-guidance-names.py` docstring.
Do: widen line 2 to "Fail when a skill or Cursor rule names code that no longer exists, or a kit folder has no DesignSystem README row." (wrapped to the file's width). Move the kit paragraph ("It also fails on a folder under …") to after the "It cannot see a renamed short or generic name…" paragraph, just before the usage line.
Check: `python3 scripts/check-guidance-names.py` → `check-guidance-names: OK`.

**F-4 — Drop the not-yet-landed clause** (from MNT-3)
Files: `docs/matching-model.md` line 3.
Do: change `this page is the "what", kept true as the layers land.` to `this page is the "what".`
Check: `rg -n 'as the layers land' docs/matching-model.md` → no hit.

Check for the round: scripts/check-pr.sh PASS.

## Round 1 — fix
Head `1f5c26d6`. Gate PASS (orchestrator run). F-1 – F-4 done, one commit each. No reviewer goes to round 2 (no Fix-now blocker or major). The orchestrator read the fix commits:
- CON-1 **Resolved**: the plan's line 4 is a `**Landed:**` note, and `check-plan-docs.py` reports 0 errors.
- MNT-1 **Resolved**: the line 73 enum comment names `core/database/deleteimpact/kinds.go`; the AC-2.4 grep still gives one hit.
- MNT-2 **Resolved**: the docstring summary covers both checks, and the kit paragraph follows the "cannot see" paragraph.
- MNT-3 **Resolved**: "as the layers land" is gone.

## Result
Gate PASS · CI green · open: none · follow-ups: none · to user: — · CON miss: none
