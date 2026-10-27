# Ship log — S11-43 Lower the candidate gate

**Plan:** [`S11-43-candidate-gate.md`](S11-43-candidate-gate.md) · **Branch:** `s11-43-candidate-gate` · **PR:** #588
**State:** ready

## Round 0 — implement
Head `1a3cb215`. Gate PASS (orchestrator run). PR #588 (draft). Report deviations: T-1/T-2 write `vocabkey.PropertyKey(vocabkey.PropertyName)` (the constant is untyped, as `TestPartialNamesAreWeak` already does); T-1's gate assertion uses `t.Errorf` so STEP-1's three red assertions all run. Both in the plan's Landed note. AC-1.5 (§10 manual check) left to the user.

## Round 1 — review
Reviewers: CON (Sonnet), MNT, SEC+PRF. Skipped: UX (no `macos/App/**`, strings, or error codes). SEC: `core/matchregistry`; PRF: `core/graphalign` (graph code).

| ID | Lens | Severity | Title | Decision | Evidence / reason |
| --- | --- | --- | --- | --- | --- |
| MNT-1 | MNT | major | `namevaluetest.Given` added while four private given-only builders remain in three packages | Fix now → F-1 (same-pass cleanup; supersedes the plan's Out line "Moving the private given-only builders") | verified by body: `namealiases/registry_test.go:14` `givenName`, `namealiases/namealiases_test.go:356` `given`, `valuerules/valuerules_test.go:17` `given`, `catalogsession/alias_cache_test.go:73` `aliasGiven` each build one `PartTypeGiven` part; no call site passes multi-word text (`rg` over string-literal calls with a space is empty); test-only, nothing a researcher sees |
| MNT-2 | MNT | minor | New graphalign tests build the given-only name property by hand with a cast | Fix now → F-2 | verified: `align_test.go:615`, `candidates_test.go:106` repeat `vals(vocabkey.PropertyKey(vocabkey.PropertyName), nameOf(namevaluetest.Given("Amy")))`; `nameProp` (`align_test.go:1871`) is the typed sibling; lines this PR adds |
| MNT-3 | MNT | minor | `add-name-value` calls `namevaluetest` the "Western fixture helper" | Fix now → F-3 | verified: `.agents/skills/add-name-value/SKILL.md:22` |
| MNT-4 | MNT | minor | `nameVals` / `surnameVals` in `align_test.go` hand-build what `Western` builds | Fix now → F-4 (same-pass cleanup: file this PR changes, copies of the fixture builder) | verified: `align_test.go:587-605` part lists and Form match `namevaluetest.Western` |
| SEC/PRF | SEC+PRF | — | No findings | — | — |
| CON | CON | — | No findings | — | — |

Nits (logged): `MNT-5 · core/namevalue/namevaluetest/namevaluetest.go:26 · wrap Given's doc comment to match Western's (~78 cols)` · `CON-1 · ship log untracked (by design: committed at finish)`

### Fix plan 1
Supersedes the plan's Out line "Moving the private given-only builders … Not planned here" (same-pass cleanup); note it in the plan's Landed note.

**F-1 — Private given-only builders move onto `namevaluetest.Given`** (from MNT-1)
Files: `core/database/namealiases/registry_test.go`, `core/database/namealiases/namealiases_test.go`, `core/database/valuerules/valuerules_test.go`, `core/catalogsession/alias_cache_test.go`, the plan (Landed note).
Failing test first: none new — the existing tests in `namealiases`, `valuerules`, `catalogsession` must pass with no expected value changed.
Do: in each wrapper, replace the inner `&namevalue.Value{Parts: …PartTypeGiven…}` literal with `namevaluetest.Given(text)`; keep each wrapper's `Type` (and `valuerules`' `NamePattern`). Merge `namealiases`' `givenName` and `given` into one helper (keep `given`, rename call sites of `givenName`). Drop imports that go unused.
Check: `rg -nU "PartTypeGiven\}\}" core --glob '*_test.go'` and `rg -n "Type: namevalue.PartTypeGiven,\s*$" core --glob '*_test.go'` print no wrapper bodies in those four files; `CGO_ENABLED=1 go test -tags fts5 ./core/database/namealiases/ ./core/database/valuerules/ ./core/catalogsession/` passes; scripts/check-pr.sh PASS.

**F-2 — Typed given-only name property for graphalign tests** (from MNT-2)
Files: `core/propertycompare/propertycomparetest/propertycomparetest.go`, `core/graphalign/align_test.go`, `core/graphalign/candidates_test.go`. Skill: `add-match-rule`.
Failing test first: none — T-1 and T-2 pass unchanged in what they assert.
Do: add `propertycomparetest.Given(form)` beside `Name`, wrapping `namevaluetest.Given`, with a doc comment in `Name`'s style. Add `givenProp(form)` beside `nameProp` (`align_test.go:1871`) returning `(vocabkey.PropertyKey, propertycompare.Value)`. T-1 and T-2 build their name with `vals(givenProp("Amy"))` (and the other names likewise via `nameProp`/`givenProp`), no `vocabkey.PropertyKey(...)` cast. Update the plan's Landed note deviation 1 (the cast is gone).
Check: `rg -n "vocabkey.PropertyKey\(vocabkey.PropertyName\), nameOf\(namevaluetest" core/graphalign` prints nothing; T-1, T-2 pass.

**F-3 — `add-name-value` names both fixture builders** (from MNT-3)
Files: `.agents/skills/add-name-value/SKILL.md`.
Do: line 22 becomes `namevaluetest/             # Western and Given fixture helpers`. Also wrap `Given`'s doc comment in `namevaluetest.go` to match `Western`'s width (nit MNT-5).
Check: scripts/check-pr.sh PASS.

**F-4 — `nameVals` / `surnameVals` build through the shared fixture** (from MNT-4)
Files: `core/graphalign/align_test.go`.
Failing test first: none — every test using them passes unchanged.
Do: `nameVals(given, surname)` returns `vals(nameProp(given + " " + surname))` (or `vals(vocabkey.PropertyKey(...), propertycomparetest.Name(...))` if `vals` shape needs it); `surnameVals(s)` returns `vals(nameProp(s))`. Keep the function names and their doc comments; check that `Western` reads a single word as a surname (it does for `surnameVals`' current shape) and that no caller passes a multi-word given name to `nameVals` — if one does, stop and report.
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/graphalign/` passes; scripts/check-pr.sh PASS.

## Round 1 — fix
Head `17928c3f`. Gate PASS (orchestrator run). F-1 – F-4 done in three commits (F-2 and F-4 share `align_test.go`). Report deviations: no `completed.md` change (round is test- and doc-only); plan's Landed note rewritten (cast gone, Out line on private builders marked superseded, `nameVals`/`surnameVals` noted). Check greps print nothing.
Orchestrator read: MNT-2 **Resolved** (`givenProp` beside `nameProp`, `propertycomparetest.Given` beside `Name`; cast gone). MNT-3 **Resolved** (SKILL.md:22). MNT-4 **Resolved** (`nameVals`/`surnameVals` via `nameProp`). MNT-5 nit done.

## Round 2 — review
Reviewers: MNT (MNT-1 was a Fix-now major). Scope `1a3cb215...17928c3f`.

| ID | Lens | Severity | Title | Decision | Evidence / reason |
| --- | --- | --- | --- | --- | --- |
| MNT-1 | MNT | major | (round 1) private given-only builders | **Resolved** | `git grep PartTypeGiven` leaves no given-only wrapper in the four files; packages pass |

No new findings. No Fix plan 2.

## Result
Gate PASS · CI pending · open: none · follow-ups: none · to user: — · CON miss: none
