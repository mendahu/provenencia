# Ship log — S11-51 One person's links for every reader

**Plan:** [`S11-51-person-links.md`](S11-51-person-links.md) · **Branch:** `s11-51-person-links` · **PR:** #575
**State:** ready

## Round 0 — implement
Head `b9663679`. Gate PASS (orchestrator run). PR #575 (draft). Report deviations: T-1 written with `Relationships` in one commit (not watched red first); no Landed note (code matches the plan); manual check left to the user.

## Round 1 — review
Reviewers: CON (Sonnet), MNT, SEC+PRF. Skipped: UX (no `macos/App/**`, strings, or error codes).

| ID | Lens | Severity | Title | Decision | Evidence / reason |
| --- | --- | --- | --- | --- | --- |
| CON-1 | CON | major | Plan has no Landed note, which the DoD requires | Fix now → F-1 | verified: the plan has only the Status line; DoD "a Landed note lists every difference from the plan" |
| MNT-1 | MNT | minor | Edge deep copy written in `LinkSet.Add` and twice in `TestRelationships` | Fix now → F-2 | same-pass cleanup: this PR adds the two test copies (`reader_test.go:388`, `:402`); `kinship` has `Shape.Clone` and `Recording.Clone` but no `Edge.Clone` |
| CON-2 / MNT-2 | CON, MNT | minor | `loadCanon` / `loadLayerBridges` doc comments not re-wrapped after the rename | Fix now → F-3 | verified: `canon.go:40`, `layer.go:157` run past the paragraph's width; files this PR changes |

Nits (logged): `CON-3 · core/kinship/kinship.go:5 · backticks around LinkSet in the package doc (plan D-1 specified them)` · `MNT-3 · docs/ideas/kinship-view.md:9 · person pages could name Reader.Relationships, not Reader.For` · `MNT-4 · core/kinship/links.go · PlaceImplied doc could say callers append new links in implied order`
SEC+PRF: no findings.

### Fix plan 1

**F-1 — Landed note on the plan** (from CON-1)
Files: `docs/deployment-plan/spike-11/prs/S11-51-person-links.md`.
Do: add a `**Landed:**` line in the header (after **Planned:**, as S11-39's plan does) listing each difference from the plan: T-1 written in the same commit as `Relationships`, not watched red first; `TestLinkSet` gained the subtest "an added implied link records as Implied.Recording" (planned in T-2); this fix round's `Edge.Clone` (F-2). Add the evidence: `PROVENENCIA_GRAPH_VERIFY=1` over `canonlinks` passed; `TestLargeSourceBudgets -count=3` passed with budgets unchanged.
Check: `python3 scripts/check-plan-docs.py docs/deployment-plan/spike-11/prs/S11-51-person-links.md` reports no errors.

**F-2 — `Edge.Clone`** (from MNT-1)
Files: `core/kinship/kinship.go`, `core/database/canonlinks/reader_test.go`, `core/kinship/kinship_test.go` or `links_test.go`.
Failing test first: a `kinship` test that `Edge.Clone` returns an edge equal to the input that shares no bytes of `A`, `B` or any association (overwrite the input's bytes, the clone is unchanged).
Do: `func (e Edge) Clone() Edge` beside `Shape.Clone` in `kinship.go` (ends and every association copied; keys copied by value). `TestRelationships`' `cloneLinks` becomes `Link{Edge: l.Edge.Clone(), Recording: l.Recording.Clone()}`; `cloneImplied` uses `l.Edge.Clone()` plus its `Event`, `Through` and `From` copies. `LinkSet.Add` keeps its own loop (it fills the association set in the same pass). `canon.go`'s `CanonEdge` build is not an `Edge` and stays.
Check: `grep -rn 'bytes.Clone(l.A)' core` prints only `core/kinship/links.go:68`; `CGO_ENABLED=1 go test -tags fts5 ./core/kinship/ ./core/database/canonlinks/` passes; scripts/check-pr.sh PASS.

**F-3 — Re-wrap two doc comments** (from CON-2, MNT-2)
Files: `core/database/promotealign/canon.go` (`loadCanon` comment), `core/database/promotealign/layer.go` (`loadLayerBridges` comment).
Do: re-wrap each paragraph to its neighbors' width (about 78 columns); no word changes.
Check: gofmt clean; scripts/check-pr.sh PASS.

## Round 1 — fix
Head `1006fbdb`. Gate PASS (orchestrator run), plus `PROVENENCIA_GRAPH_VERIFY=1` over `canonlinks` and `TestLargeSourceBudgets -count=3` (orchestrator run, both pass). F-1 – F-3 done; `TestEdgeClone` failed first. Deviation: `cloneImplied` also deep-copies `From`.
MNT-1 **Resolved** (`Edge.Clone` beside `Shape.Clone`; the grep prints only `links.go:68`). CON-2 / MNT-2 **Resolved** (both comments re-wrapped, no word changes).

## Round 2 — review
Reviewers: CON (had a Fix-now major). MNT, SEC+PRF: no Fix-now blocker or major.
CON-1 **Resolved** (Landed line at plan line 16 lists every difference and the evidence). CON-2 **Resolved**. No new findings.

## Result
Gate PASS · CI green · open: none · follow-ups: MNT-3 (nit: `docs/ideas/kinship-view.md` could name `Reader.Relationships`) · to user: — · CON miss: none
