# Ship log — S10-50 Promote sheet lines show the values they compared

**Plan:** [`S10-50-promote-sheet-values.md`](S10-50-promote-sheet-values.md) · **Branch:** `s10-50-promote-sheet-values` · **PR:** #413
**State:** ready

## Round 0 — implement
Head `407f96db`. Gate PASS (orchestrator run, against `ad7f5ac8`). Report deviations:
- `TestExhibitNameWithoutForm` blanks `name_values.form` by SQL after filing, because `namevalues.validate` rejects an empty Form, so the writers cannot build the plan's fixture. Recorded in the plan's Landed note.
- One flaky `xcode-test` failure (`WorkspaceSessionTests.invalidateAllMatching()`) on a docs-only commit; passed on re-run without changes.
- Manual check (§10) not run in the app.

## Round 1 — review
Head `407f96db`. Verdicts: CON clean · MNT clean · UX clean · SEC clean · PRF clean. No blocker or major, so no fix round and no round 2 (minors are fixed now only when a fix plan already touches the file).

| ID | Lens | Severity | Title | Decision | Evidence / reason |
| --- | --- | --- | --- | --- | --- |
| CON-1 | CON | minor | T-9's `event_type` check passes if the line is missing | Follow-up | test strength only; `event_type` is a profile Feature (`registry.go:276`) |
| CON-2 | CON | minor | `TestOrderIgnoresTermLabel`, `TestValuesCarryTermLabel` not table-driven | Follow-up | `.cursor/rules/go-tests.mdc` |
| CON-3 | CON | minor | Manual check not run in the app | Logged | the user runs §10 before merge; PR body says so |
| MNT-1 | MNT | minor | `attachExhibits` doc comment still says "fills pin lines on handle rows" | Follow-up | `exhibits.go:28` |
| MNT-2 | MNT | minor | `Value.Label` doc names only Reasons, not the sheet | Follow-up | `pairs.go:99` |
| MNT-3 | MNT | minor | `event_titles.go` copies `VocabTerm` into an anonymous struct for `FromStored` | Follow-up | D-2 kept the signature on purpose |
| MNT-4 | MNT | nit | `sortKey` restates `Label`'s case order | Logged | |
| MNT-5 | MNT | nit | `withRowLines` holds two rules in one if/else | Logged | |
| UX-1 | UX | minor | On a no-exhibit row the hint reads "0 pinned comparisons" above a drafted argument; the string has no plural variants | Follow-up | Swift and copy change; plan Out "Any Swift change" (D-7) |
| UX-note | UX→CON | minor | A ref line's Swift id is `"||" + propertyKey`; two Properties with one key and different origins would share it | Follow-up | predates the PR (`GenealogyStore.swift:257`); CON found ids unique for product Properties |
| PRF | PRF | — | — | — | clean; `BenchmarkPairValues/plain` 0 allocs/op |
| SEC | SEC | — | — | — | clean |

## Result
Gate PASS (`407f96db`) · CI pending · open: none · follow-ups: UX-1, UX-note, MNT-1, MNT-2, MNT-3, CON-1, CON-2 · to user: —
