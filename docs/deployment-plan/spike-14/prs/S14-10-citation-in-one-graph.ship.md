# Ship log — S14-10 A Citation belongs to one Evidence graph

**Plan:** [`S14-10-citation-in-one-graph.md`](S14-10-citation-in-one-graph.md) · **Branch:** `s14-10-citation-in-one-graph` · **PR:** #645
**State:** ready

## Round 0 — implement

Head `8497d93d`. Gate PASS (orchestrator re-ran `scripts/check-pr.sh`). Report deviations: migration is `000059.sql` (S13-25 took 000058); T-1/T-2 are subtests of `TestCitations`; commits grouped by layer; `FakeStore.citationCountsGraphIDs` recorder added for `WorkspaceSessionTests.citationReadsRefreshOnlyTheirGraph`; D-11 names map bullets absent from the Spike 14 map; one flaky `SourceGraphFormModelTests` xcode-test run (passed on rerun, unchanged). §10 manual check (AC-3.1) not run.

## Round 1 — review

Reviewers: CON (Sonnet), MNT, UX, SEC+PRF. Skipped: none (diff touches `macos/App/**`, Localizable, `core/apperr`, migrations, `api/ffi/**`, `core/database/**`).

| ID | Lens | Severity | Title | Decision | Evidence / reason |
| --- | --- | --- | --- | --- | --- |
| UX-1 | UX | major | Graph delete confirm lists at most 20 Citation refs, no total | Fix now → F-1 | verified: `PVDeleteImpact.cascadeLines` builds from `group.listed` only and never reads `group.total`; `listedCap = 20`. Same gap in the `identity_claims.subject_id` and default lines, fixed in the same function |
| MNT-1 | MNT | major | Interpretation data model keeps the old isolation wording and a `citations` block without `graph_id` | Fix now → F-2 | verified: `interpretation-layer-data-model.md:579` and invariant 3 (`:937`) still say "Citation Artifact share one Evidence graph" |
| MNT-2 | MNT | major | add-catalog-delete template calls `ReleaseCitation` then still says DELETE parent + parent change | Fix now → F-3 | verified: `SKILL.md:77-82` |
| MNT-3 | MNT | minor | `releaseCitation` hand-lists Citation columns; `citations.Delete` reads the row twice | Fix now → F-4 | minor in code this PR adds; column-parity test guards the copy the import cycle forces |
| MNT-4 | MNT | nit | `citations.go:395` comment "Notes CASCADE." is stale | Fix now → F-4 | stale comment in a changed file |
| UX-2 | UX | minor | Save after another window unlinks the Artifact shows generic "That citation isn't valid." | Follow-up | D-2 chose `citations.invalid` and rejected a user-kind code for "not linked"; reopening that is a Decision change. A stale `artifacts` pre-check would not see another window's unlink. PR body follow-up |
| SEC-note | SEC+PRF (outside lens) | — | Graph delete of a migrated split Citation fails with a raw FK error | Rejected | Plan §11 Risks accepts it: no dogfood Citation has Observations in two graphs (query 2026-10-29, D-1) |

Nits (logged): `UX-3 · L10n.swift:8336 · singular "citations" with one ref (folded into F-1)` · `MNT-5 · evidencegraphs.go:509-540 · queryCitationRows generic over one query` · `MNT-6 · facets.go · ReleaseCitation/ReleaseClaim identical wrappers; helper if a third appears`

CON: no findings. SEC+PRF: no findings.

### Fix plan 1

**F-1 — The graph delete confirm counts every Citation it deletes** (from UX-1, UX-3)
Files: `macos/App/DesignSystem/Recipes/DeleteImpact/PVDeleteImpact.swift`, `macos/App/Platform/L10n.swift`, `macos/App/Resources/Localizable.xcstrings`, `macos/ProvenenciaTests/PVDeleteImpactTests.swift`. Skill: `add-localized-string`, `add-swift-test`.
Failing test first: `PVDeleteImpactTests` — a `citations.graph_id` cascade group with `total: 25` and 20 listed refs; the line must say 5 more (or carry 25). Second case: one Citation reads in the singular.
Do: in `cascadeLines`, when `group.total > refs.count`, append the remainder to the joined refs with the existing `L10n.DeleteImpact.summaryMore(count:)` (or an equivalent existing overflow string), for every case in the switch (handle, citations, default), so no cascade line understates. Make `deleteImpact.confirm.deletesCitations` a plural variant on `group.total` (one / other) so one Citation reads singular. Keep D-10's wording for the plural.
Check: the new tests pass; scripts/check-pr.sh PASS.

**F-2 — The Interpretation data model teaches the new rule** (from MNT-1)
Files: `docs/interpretation-layer-data-model.md`.
Do: change line ~579 and invariant 3 (~937) to rule 24's wording (the Subject and Subject value are in the Citation's graph; reification exception unchanged). Add `graph_id BLOB NOT NULL REFERENCES evidence_graphs(id) ON DELETE CASCADE` to the `CREATE TABLE citations` block after `artifact_id`, plus one paragraph in the style of the `subjects` `graph_id` paragraph: required on create, linked to its Artifact, never changes, an audited facet of the graph (a graph delete releases its Citations).
Check: `rg -n "Citation Artifact" docs/interpretation-layer-data-model.md` prints nothing; scripts/check-pr.sh PASS.

**F-3 — The add-catalog-delete template stops deleting the row twice** (from MNT-2)
Files: `.agents/skills/add-catalog-delete/SKILL.md`.
Do: in the § Domain `Delete` example, put back `changes, err := deleteimpact.ReleaseFacets(tx, catalogmodel.KindX, id)` as the generic line; add one comment line after it: a kind with a one-row release (`ReleaseClaim`, `ReleaseCitation`) calls that instead, which deletes the row and records its change, and skips "DELETE parent" and the parent change. Keep the bullet at line ~93. Avoid literal `releaseX`/`ReleaseX` placeholders (check-guidance-names rejects them).
Check: scripts/check-pr.sh PASS (guidance-names).

**F-4 — Guard the Citation column copy; one read in `citations.Delete`** (from MNT-3, MNT-4)
Files: `core/database/citations/citations.go`, `core/database/deleteimpact/facets.go` or its test file.
Failing test first: a test that creates a Citation, calls `deleteimpact.ReleaseCitation`, and asserts the delete change's field keys equal the column names from `PRAGMA table_info(citations)`. (Passes today; it guards the next column. Write it, confirm it passes, then make sure it fails if you drop one column from `releaseCitation`'s select locally — do not commit that.)
Do: in `citations.Delete`, make the existence check `SELECT 1 FROM citations WHERE id = ?` (missing → `ErrInvalid` as now) instead of the full `scanOne(sqlGet)`. Reword the `citations.go:~395` comment: its notes are released, audited, through `ReleaseCitation`.
Check: `go test ./core/database/citations/ ./core/database/deleteimpact/` pass; scripts/check-pr.sh PASS.

## Round 1 — fix

Head `efeb4396`. Gate PASS (orchestrator re-ran `scripts/check-pr.sh`). F-1 to F-4 done. Deviations: the plural `deletesCitations` carries the count ("The graph’s %1$lld citations are deleted with it: %2$@.", one: "citation is") because a plural variation must reference its argument; overflow lines join refs with ", " then ", and N more". MNT-3 and MNT-4 Resolved (orchestrator read `38ab07c6`: `SELECT 1` existence check, reworded comment, `TestReleaseCitationChangeNamesEveryColumn` compares change keys with `PRAGMA table_info`).

## Round 2 — review

Reviewers: UX (UX-1), MNT (MNT-1, MNT-2). Scope `8497d93d...efeb4396`. Not re-sent: CON, SEC+PRF (no findings in round 1).

UX-1 Resolved (UX round 2: true total, ", and N more" past the 20-ref cap on every cascade line, singular for one). MNT-1 and MNT-2 Resolved (MNT round 2). No new blockers. No Fix plan 2.

## Result

Gate PASS · CI pending · open: none · follow-ups: UX-2 (specific copy when a save races another window's unlink), MNT-5, MNT-6 (nits) · to user: §10 manual check · CON miss: none
