# Ship log — S10-23 Sex / gender on the Person page

**Plan:** [`S10-23-person-gender.md`](S10-23-person-gender.md) · **Branch:** `s10-23-person-gender` · **PR:** #407
**State:** review 2

Base: `main` @ `cd370713` (plan Planned against `1fc683b`; no `macos/`, `core/`, or `api/` changes since).

## Round 0 — implement
Head `c563a5ff`. Orchestrator gate PASS (against `cd370713`). Deviations: none in behavior; `sexAtBirthBadge` copy written with the treatment file (no separate STEP-3); `colors:` case added to the `PVThumbnail` preview (allowed by STEP-1). Manual check (§10) not run: needs the running app. One flaky `WorkspaceSessionTests.invalidateAllMatching` failure on the implementer's run, passed on rerun, outside this PR.


## Round 1 — review
Head `c563a5ff`. Performance and Security: clean, no findings. UX: clean. Maintainability: clean. Conformance: changes needed (CON-1).

| ID | Lens | Severity | Title | Decision | Evidence / reason |
| --- | --- | --- | --- | --- | --- |
| CON-1 | CON | major | Manual check §10 not run | To user | The plan and skill give the manual check to the user; it needs the running app and VoiceOver. Goes in the report. The Badge label is `.accessibilityLabel` on a single-`Text` pill, so the risk of VoiceOver reading the raw text is low; if the user's step 1 shows it, the fix is one modifier |
| CON-2 | CON | nit | Landed note gets the commit split wrong | Fix now → F-2 | verified with `git show --stat c6297d11`; docs only, same file as F-2 |
| MNT-1 | MNT | minor | Rule and kit README omit term keys and `colors:` | Fix now → F-1 | verified: `macos-client.mdc:20` lists no term keys; the fix is two lines of docs |
| MNT-2 | MNT | minor | Go subtest repeats catalog setup, hard-codes its table | Fix now → F-3 | verified: setup block is a line-for-line copy of the sibling subtest |
| MNT-3 | MNT | nit | `gender(_:)` / `sex(_:)` test helpers near-copies | Logged | Nit; F-3 does not touch that file |
| UX-1 | UX | minor | Long researcher-term label in the Badge cannot truncate | Follow-up | Seeded labels are short; manual check with a long term will confirm |
| UX-2 | UX | nit | VoiceOver heading navigation skips the Badge | Logged | Taste; the Gender row still speaks the value |

### Fix plan 1
**F-1 — Docs name the term keys and the `colors:` parameter** (from MNT-1)
Files: `.cursor/rules/macos-client.mdc`, `macos/App/DesignSystem/README.md`.
Failing test first: none (docs).
Do: add "seeded term keys" to the `ConclusionVocabulary.swift` parenthetical at `macos-client.mdc:20`. Append to the README Thumbnail row: "optional `colors:` (tint and line) on mark and symbol states (S10-23)".
Check: scripts/check-pr.sh PASS.

**F-2 — Landed note commit split is accurate** (from CON-2)
Files: `docs/deployment-plan/spike-10/prs/S10-23-person-gender.md` (Landed note only; no Decisions, Scope, or ACs).
Failing test first: none (docs).
Do: reword the first bullet: "STEP-2's header structs and treatment, STEP-3's copy, and STEP-4 landed in one commit; the term keys and their Go subtest are their own commit, and the thumb colours and tokens another."
Check: `python3 scripts/check-plan-docs.py` on the plan reports 0 errors.

**F-3 — One catalog helper for the vocabulary subtests** (from MNT-2)
Files: `api/ffi/swift_vocabulary_test.go`.
Failing test first: not applicable; behaviour is unchanged. `TestSwiftVocabularyMatchesGo` stays green.
Do: extract the `database.Create` + `t.Cleanup` + `subjectvocab.Install` block into one `installedCatalog(t)` helper beside `swiftEnumBody`, used by "seeded Property keys are seeded" and "seeded term keys are seeded". Optionally fail when an `enum Seeded\w+Term` in `ConclusionVocabulary.swift` has no row in the term table.
Check: `CGO_ENABLED=1 go test -tags fts5 ./api/ffi -run TestSwiftVocabularyMatchesGo` passes; scripts/check-pr.sh PASS.

## Round 1 — fix
Head `d553d4c5`. Orchestrator gate PASS. F-1, F-2, F-3 done. Optional F-3 scan (every `Seeded…Term` has a table row) not added.
