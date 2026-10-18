# Ship log — S11-06 Migrate the reconciler onto value compare

**Plan:** [`S11-06-migrate-reconciler.md`](S11-06-migrate-reconciler.md) · **Branch:** `s11-06-migrate-reconciler` · **PR:** #456
**State:** fixing 1

## Round 0 — implement
Head `f6ea8808`. Gate PASS (orchestrator run). Report: no deviations from Decisions or Scope; extra test rows beyond T-1–T-7 (listed in the Landed note); manual check §10 steps 3–4 not run by hand.

## Round 1 — review
| ID | Lens | Severity | Title | Decision | Evidence / reason |
| --- | --- | --- | --- | --- | --- |
| MNT-1 | MNT | major | Name module pairs words order-free to group/fold but by position to outvote | To user | verified by reading: `outvotes` indexes `parts[i]` (`names.go:124-138`) while `split`/`fold` use value compare (order-free). `[James Kenneth]` vs `[Kenneth Jmaes]` is not outvoted. Fix contradicts D-4 "`outvotes`' rule … unchanged" |
| MNT-2 | MNT | major | Reconciler builds its own particle set beside `splitText` | To user | verified: `Particles` read in `valuecompare/text.go:12` and `autoreconcile/modules.go:69-71`; `SplitWords` had no callers at base. Fix contradicts D-4 (`particles` field) and touches `core/valuecompare` (stop condition) |
| CON-1 | CON | minor | `laterStart`/`earlierEnd` branches untested | Fix now (rides along) | new logic in a test file this PR adds; few lines |
| CON-2 | CON | minor | `TestDateWindow` not table-driven; no one-sided-range row | Fix now (rides along) | go-tests.mdc; test file this PR adds |
| MNT-4 + CON-3 | MNT, CON | minor | §7.2 "ordered list of words"; §12 "subsumption"; test group label | Fix now (rides along) | `rg subsumption` hits `conclusion-reconciliation.md:279`, `names_test.go` |
| MNT-3 | MNT | minor | Third copy of date test builders | Follow-up | D-10 chose the copy; extract `datevaluetest` later |
| UX-1 | UX | nit | Same-day tie keeps the less precise date | Logged | D-7 says a tie keeps `a`; follow-up idea: prefer the narrower window |
| SEC (routed to CON) | SEC | minor | `DateEnds` doesn't reject an unknown qualifier | Logged | `DateEnds` is S11-04's, unchanged here; `datevalue.Validate` rejects unknown qualifiers on write |
| PRF | PRF | — | ~1.46× slower per name group (benchmark) | — | Clean: same growth, off main thread |

Paused for the user on MNT-1 and MNT-2.

**User, 2026-10-17:** MNT-1 → Fix now; MNT-2 → Fix now. Both amend D-4; the amendment is recorded in the plan (F-1, F-2) with a pointer to this log.

### Fix plan 1

**F-1 — Outvote over value compare's word pairing** (from MNT-1)
Files: `core/autoreconcile/names.go`, `core/autoreconcile/names_test.go`. Skill: `add-match-rule`.
Failing test first: `TestReconcileNamesWords` new row — given `[James Kenneth]` from two sources and `[Kenneth Jmaes]` from one; the result's given is `[James, Kenneth]` only (the minority is outvoted). Today it shows both.
Do: rewrite `nameModule.outvotes(winner, other)` to call `valuecompare.Compare(other.cmp, winner.cmp, m.rules)` (the stored `nameUnit.cmp`, as `fold` does). Outvote when `det.OnlyA` and `det.OnlyB` are both empty and, for every `p` in `det.Parts`, `p.A.Text == p.B.Text` or `valuecompare.SpellingSimilarity(p.A.Text, p.B.Text, m.rules.Words.FuzzyFloor, m.rules.Words.ShortVariantMinLength) > 0`. The per-word test is exactly today's; only the pairing changes. Keep the doc comment's rule and say pairing comes from value compare. Existing outvote rows (Robbins/Robins outvoted; Jake/James and Kevin/Kenneth never) must still pass unchanged. If `nameUnit.parts` has no reader left, delete the field.
Also amend the plan's D-4 cell: replace "`outvotes`' rule (per-word `SpellingSimilarity`, same word count) and `assemble`'s logic are unchanged" with "`outvotes` keeps its per-word `SpellingSimilarity` test but pairs words through `Compare`'s `Detail.Parts` with no unpaired word (amended 2026-10-17 by the user, ship log round 1 MNT-1); `assemble`'s logic is unchanged", and update the §2/§3 row "The name module's spelling-only outvote rewritten on relations | not planned (D-4)" only if it now reads false (it stays true: the test is not rewritten on relations).
Check: the new row passes; `go test -tags fts5 ./core/autoreconcile/...`; scripts/check-pr.sh PASS.

**F-2 — One particle split** (from MNT-2)
Files: `core/valuecompare/words.go`, `core/valuecompare/text.go` (if `splitText` becomes a call to it), any `SplitWords` test in `core/valuecompare`, `core/autoreconcile/modules.go`, `core/autoreconcile/names.go`. Skill: `add-match-rule`.
Failing test first: none needed (refactor); existing `SplitWords` / name tests are the guard.
Do: change `valuecompare.SplitWords(s string, particles map[string]bool)` to `SplitWords(s string, w WordRules) []string`, implemented as `splitText(s, w)` (or make `splitText` the exported one and delete the other — one function, one particle test). Update its doc. `nameModule.readName` calls `valuecompare.SplitWords(p.Value, m.rules.Words)`. Delete `nameModule.particles`, its doc mention, and the map loop in `moduleFor` (`return nameModule{rules: rules}`). Update any test call of `SplitWords`. Amend the plan's D-4 cell: replace the `particles` sentence with "builds each part's words with `valuecompare.SplitWords(p.Value, m.rules.Words)`, the split value compare's text and name comparers use (amended 2026-10-17 by the user, ship log round 1 MNT-2)". The plan's stop condition on `core/valuecompare` changes does not apply to this signature change; the user approved it.
Check: `rg -n 'Particles' core --glob '!*_test.go'` shows only `core/valuecompare` (and `core/matchregistry`); `go test -tags fts5 ./core/...`; scripts/check-pr.sh PASS.

**F-3 — Tests for the by-day membership branches** (from CON-1)
Files: `core/database/conclusionheaders/placechains_internal_test.go`.
Do: add rows to `TestMembershipDatesByDay`, each run in both argument orders where the result is order-independent: phrase-only start vs `1 SEP 1850` → `1 SEP 1850`; `BEF 1850` start vs `1840` → `1840`; `AFT 1867` end vs `1 JUL 1867` → `1 JUL 1867`; tie `1850` vs `1850` → argument `a` (one order only, asserting identity of `a`).
Check: `go test -tags fts5 ./core/database/conclusionheaders/...`.

**F-4 — `TestDateWindow` as a table, with one-sided ranges** (from CON-2)
Files: `core/valuecompare/datewindow_test.go`.
Do: rewrite as a table with `t.Run` per case, same seven cases, plus `KindRange` with only `EndYear` 1900 → `Lo` nil, `Hi` set; only `StartYear` 1885 → `Lo` set, `Hi` nil.
Check: `go test ./core/valuecompare/...`.

**F-5 — Retire "ordered" and "subsumption" wording** (from MNT-4, CON-3)
Files: `docs/conclusion-reconciliation.md` (§7.2 unit bullet; §12 row ~L279), `core/autoreconcile/names_test.go` (group labels).
Do: "its ordered list of **words**" → "its **words**"; §12 "subsumption" → "containment per part type"; rename the `"subsumption"` test group labels in `TestReconcileNames` to `"fold"`.
Check: `rg -n 'subsumption' --glob '!docs/deployment-plan/**' .` prints nothing; scripts/check-pr.sh PASS.
