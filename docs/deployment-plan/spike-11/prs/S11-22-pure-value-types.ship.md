# Ship log — S11-22 Pure value types

**Plan:** [`S11-22-pure-value-types.md`](S11-22-pure-value-types.md) · **Branch:** `s11-22-pure-value-types` · **PR:** #434
**State:** ready

Worktree: `../provenencia-s11-22`, branched from `origin/main` @ `233e87ad` (plan's base `af50bd6`).

## Round 0 — implement
Head `233e87ad` (no commits). Gate not run.
**Stopped on:** base check. §4 "Property constants" is stale: `properties.go` gained `ValueTypeDepiction` (S10 depiction work, L46–48), `valueTypeOK` accepts it, and a new `researcherCreatable` (L480) exists. The plan names six value types; it does not say whether `ValueTypeDepiction` moves (D-6's regex would flag it if it stays) or where `researcherCreatable` lives. Verified by the orchestrator against the file. Taken to the user.
**User decision (2026-10-17):** `ValueTypeDepiction` moves to `propertymeta` (nine constants); `researcherCreatable` stays in `properties`. Plan and review log amended in `2b48eb75`. Round 0 resumed.

Head `51764575`. Gate PASS (orchestrator run). Report deviations: three §4 paths made plain text so `plan-docs` passes once the files move (`177f8a8a`, noted in Landed note); two pre-existing gofmt failures at base left alone.

## Round 1 — review
Lenses: CON (1 minor), MNT (2 minor, 1 nit), UX clean, SEC clean (1 outside-lens note), PERF clean.

| ID | Lens | Severity | Title | Decision | Evidence / reason |
| --- | --- | --- | --- | --- | --- |
| CON-1, MNT-2 | CON, MNT | minor → major | `ConclusionVocabulary.swift:89` still names `datevalues.Qualifier*` | Fix now → F-1 | verified: line 89. Re-rated: a Scope In line ("Swift doc comments name the new paths") left undelivered, not a polish item |
| MNT-1 | MNT | minor | `namevalues` package comment duplicates the type rules and points at a moved `registry.go` | Fix now → F-2 (rides along) | verified: `namevalues.go:1-15`; `datevalues.go:1-3` shows the clean shape. Comment in a file this PR changes; a few lines |
| MNT-3 | MNT | nit | `rowchange` import alone in its own group in `observations.go` | Logged | nit; fix plan does not touch the file |
| SEC-note | SEC→MNT | minor | Exported `Validate` assumes callers trimmed first; a pure-layer caller that skips trimming accepts a whitespace-only form | Follow-up | plan Out keeps trimming in the DB functions; S11-02 (first pure caller) decides whether `Validate` trims or callers must |

### Fix plan 1

**F-1 — Name the pure package in the DateQualifier doc comment** (from CON-1, MNT-2)
Files: `macos/App/Platform/ConclusionVocabulary.swift`.
Failing test first: none (comment only).
Do: L89 `datevalues.Qualifier*` → `datevalue.Qualifier*`, as `DateValueDraft.swift` L4 did. Add the file to the plan's Landed note as a fourth Swift comment §4 missed.
Check: `rg -n 'datevalues\.(Qualifier|Kind|Value)' macos` has no hits; scripts/check-pr.sh PASS.

**F-2 — Cut the namevalues package comment to the query package's own facts** (from MNT-1)
Files: `core/database/namevalues/namevalues.go`.
Failing test first: none (comment only).
Do: follow `core/database/datevalues/datevalues.go:1-3`. Keep the "lives in core/namevalue; this package holds the queries" sentence and the persistence point (Observation saves call UpdateTx so value_name_id stays stable; other callers attach a new row). Drop the type-rule paragraph, which `core/namevalue/namevalue.go` owns.
Check: `rg -n 'registry.go' core/database` has no hits; scripts/check-pr.sh PASS.

## Round 1 — fix
Head `dbae6ae2`. Gate PASS (orchestrator run). F-1 done (`a52d7749`); F-2 done (`dbae6ae2`). F-2's check pattern was loose: the one remaining `registry.go` hit is `core/database/nameformats/nameformats.go:3`, that package's own file — accepted. F-2 also dropped the "Format profiles … structured-name-model §4" line with the type paragraph — accepted.

## Round 2 — review
Lenses: CON, MNT. Scope `51764575...dbae6ae2`.
CON: clean — CON-1 resolved. MNT: clean — MNT-1, MNT-2 resolved; MNT-3 stays logged. Observation (not a finding): the dropped namevalues paragraph also said "keep DDL as open TEXT"; restore it in the query package comment next time that file is edited. No round-2 blockers, so no Fix plan 2.

## Result
Gate PASS (`dbae6ae2`) · CI pending · open: none · follow-ups: SEC-note (`Validate` assumes callers trimmed — S11-02 decides), MNT-3 (nit), "open TEXT" note · to user: — (base-check stop answered 2026-10-17)
