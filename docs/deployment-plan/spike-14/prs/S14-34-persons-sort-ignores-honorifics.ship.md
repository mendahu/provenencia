# Ship log — S14-34 Persons sort ignores honorifics

**Plan:** [`S14-34-persons-sort-ignores-honorifics.md`](S14-34-persons-sort-ignores-honorifics.md) · **Branch:** `s14-34-persons-sort-ignores-honorifics` · **PR:** #727
**State:** ready

## Round 0 — implement
Head `a35531ad`. Gate PASS (orchestrator run; base `b4556785`). Report deviations: plan Status reads "Landed in #727 (2026-11-06)" and the Context line's dogfood note link became plain text (link checker rejects a deleted target), both in the Landed note; an extra commit for that unlink; `SortKey` and `TestPersonListOrder` tests use inline `namevalue.Part` values; manual check not run. Implementer saw one `xcode-test` failure on an earlier head in `SourcePageModelTests.openRefusesAStoredScriptByItsTypeAndNamesTheFile()`, unrelated to this PR, which passed on rerun unchanged.

## Round 1 — review
Reviewers: CON, MNT, SEC+PRF (`core/database/**` matches performance). Skipped: UX (no `macos/App/**`, strings, `core/apperr`, or error codes; the only `macos/` change is `MARKETING_VERSION` in the project file).

| ID | Lens | Severity | Title | Decision | Evidence / reason |
| --- | --- | --- | --- | --- | --- |
| MNT-1 | MNT | major | `SortKey` takes an order, so it cannot read the per-profile rule §6 "Next adds" names | To user | verified: `SortKey(name, order []PartOrder)`, `persons` passes only `loaded.Profile.Sorted`; contradicts D-1 and D-5 signatures |
| MNT-2, CON-2 | MNT, CON | major | `seeded-vocabulary.md:569` calls the sorted part order the list order | Fix now | verified: file not in the diff; Scope In names this class of doc; STEP-4 missed it |
| MNT-3, CON-4 | MNT, CON | minor | `SortByTitle` and `sortParticipants` comments say Persons sort by the shown title | Fix now | verified: AC-1.1 and AC-1.2 need key and display to differ |
| MNT-4 | MNT | minor | `TestSortKey` is the third copy of the Install and Load setup | Fix now | same-pass cleanup: this PR copied it again |
| MNT-5 | MNT | minor | `SortKey` shares a name with `autoreconcile.SortKey` | To user | contradicts D-1's name; a `Profile.SortKey` method would settle it (see MNT-1) |
| MNT-6 | MNT | minor | `add-name-value` step 6 does not mention `sortKeyOmits` | Fix now | verified: no hit in `.agents`; new decision for each part type |
| MNT-7 | MNT | minor | `nameformats.Format` has no caller outside its package | To user | verified by grep: only `SortKey` and `Profile.Format`, which only tests call. Unexporting means editing Swift comments that name it, against D-6's "No Swift change" |
| CON-1 | CON | minor | AC-1.4 list case passes if a prefix-only name drops to the ref tail | Fix now | reviewer's mutation run; add a ref-only row |
| CON-3 | CON | minor | PR body says "Deviations: none" while the Landed note lists one | Fix now (orchestrator) | the PR body is rewritten at step 8; the manual check stays with the user per `ship-pr` step 8 |
| SEC-note | SEC | minor | [prefix, blank given] gives `""` and drops to the tail | Rejected | `namevalues.go:70-76,115-121` trim and `namevalue.Validate` rejects empty parts on both write paths (CON probe); only a foreign catalog could hold it, and only sort position changes |

Nits (logged): `MNT-8 · compose.go:379 · inline the one-use local order` · `MNT-9 · conclusionsummaries_test.go:970 · add closure could use f.cite and f.promoteSubject` · `MNT-10 · order.go:126 · "Persons list" and "People list" for one thing` · `MNT-11 · nameformats_test.go:308 · p and part closures are the same one-liner in two packages`

Also found in triage: `macos/App/Platform/NamePartType.swift:37` still reads "`nameformats.Format` sort key", stale after D-2. Goes with MNT-7.

SEC+PRF: no findings. `sortPersons` at 37,000 rows 864 ms → 159 ms (reviewer probe).

**User answers (2026-11-06).** MNT-1: make it `Profile.SortKey`, superseding the signatures in D-1 and D-5 → Fix now. MNT-5: closed by the receiver, no rename → Rejected. MNT-7: unexport `Format`, with a comment-only exception to D-6 → Fix now.

### Fix plan 1
**F-1 — `SortKey` becomes a method on `Profile`** (from MNT-1, MNT-5; user answer supersedes the D-1 and D-5 signatures)
Files: `core/database/nameformats/nameformats.go`, `core/database/nameformats/nameformats_test.go`, `core/database/conclusionsummaries/order.go`, `core/database/conclusionsummaries/compose.go`, `core/database/conclusionsummaries/conclusionsummaries_test.go`, `docs/structured-name-model.md`, `docs/deployment-plan/spike-14/completed.md`.
Do: `func (p Profile) SortKey(name namevalue.Value) string`, reading `p.Sorted`; the rule, `sortKeyOmits`, and "form is never a sort key" do not change. `sortPersons(rows, profile nameformats.Profile)` and `personSortKey(h, profile)`; `persons` passes `loaded.Profile`. `TestSortKey` uses the loaded profile, and `Profile{}` for the stored-order case; `TestPersonListOrder` the same, with case (c) passing `nameformats.Profile{}`. Keep `export_test.go`'s `SortPersons` var. Update every doc mention of `nameformats.SortKey` to `nameformats.Profile.SortKey`.
Check: no behavior change, every T-# passes; `grep -rn "nameformats\.SortKey" core api docs/*.md .agents` prints nothing.

**F-2 — Seed reference says what the list key is** (from MNT-2, CON-2)
Files: `docs/seeded-vocabulary.md`.
Do: line 569: the sorted style is how a list row reads; the list key is this order without `prefix` parts unless the name has nothing else (`nameformats.Profile.SortKey`, `structured-name-model.md` §4.5). The table does not change.
Check: `grep -n "Sorted style" docs/seeded-vocabulary.md` shows the new wording.

**F-3 — `Format` is private** (from MNT-7; user answer: comment-only exception to D-6)
Files: `core/database/nameformats/nameformats.go`, `.agents/skills/add-name-value/SKILL.md`, `macos/App/Platform/NamePartType.swift`, `macos/App/Platform/NameValueDisplay.swift`.
Do: rename the free function to `format`. `Profile.Format` stays exported. Skill step 6: "never in `nameformats.Profile.Format`". `NamePartType.swift:37`: Go's `nameformats.Profile.SortKey` never decorates. `NameValueDisplay.swift:81`: same order and separators as Go `nameformats.Profile.Format`. Comments only in Swift.
Check: `grep -rnE 'nameformats\.Format|[^.]\bFormat\(name' core api .agents macos/App` prints only `Profile.Format` lines; `git diff` of the two Swift files changes only `///` lines.

**F-4 — Comments say the sort title, not the shown title** (from MNT-3, CON-4)
Files: `core/database/conclusionsummaries/order.go`, `core/database/persondetail/participants.go`.
Do: `SortByTitle`'s doc: the caller supplies the row's sort title, which for a Person leaves honorifics out. `participants.go:20`: "list order of PersonsByIDs", matching `compose.go:350`. While in `order.go`, say "Persons list" in `personSortKey`'s doc (nit MNT-10).
Check: `scripts/check-pr.sh` PASS.

**F-5 — One setup helper in the `nameformats` tests** (from MNT-4)
Files: `core/database/nameformats/nameformats_test.go`.
Do: a file-local helper holding the create, `Install`, `DB`, `Load` setup; move `TestInstallLoadsWesternAndFormats`, `TestLoadFor` and `TestSortKey` onto it. Find copies by body, not name.
Check: `grep -n "Install(c)" core/database/nameformats/nameformats_test.go` prints only the helper's line, plus any test whose setup differs (say which).

**F-6 — AC-1.4 list case can fail** (from CON-1)
Files: `core/database/conclusionsummaries/conclusionsummaries_test.go`.
Failing test first: this passes on current code and must fail under a mutation. Add a ref-only row `Aaa` to the "a prefix-only name keeps its key" case; want Quinn, Rev, Aaa.
Do: with `kept = name.Parts` changed to `return ""`, confirm the case fails with Quinn, Aaa, Rev; restore.
Check: `TestPersonListOrder` passes.

**F-7 — Skill names the new decision** (from MNT-6)
Files: `.agents/skills/add-name-value/SKILL.md`.
Do: step 6 adds: decide whether the type belongs in `sortKeyOmits` (`core/database/nameformats/nameformats.go`).
Check: `scripts/check-guidance-names.py` passes.

**Landed note.** Add to the plan's Landed note: F-1 superseded the D-1 and D-5 signatures, and F-3 added comment-only Swift edits against D-6; both by the user, 2026-11-06. Do not edit the Decisions.

## Round 1 — fix
Head `6250ab1c`. Gate PASS (orchestrator run). F-1 to F-7 done. MNT-5 Rejected (`Profile.SortKey`'s receiver separates it from `autoreconcile.SortKey`, by the user's answer). Report deviations: `structured-name-model.md` and `completed.md` mentions updated with F-1 and F-3; commit `c267e9f4` alone does not compile its tests, the next commit does, and head passes.

Minors read by the orchestrator: CON-1 **Resolved** (ref-only row; implementer's mutation run fails with `[Quinn Aaa Rev]`). CON-4 and MNT-3 **Resolved** (`SortByTitle` doc, `participants.go:20`). MNT-6 **Resolved** (skill step 6). MNT-7 **Resolved** (Swift diff is two `///` lines). CON-3 is done at step 8.
Nit (logged): `order.go:19 · SortByTitle doc line runs long after F-4`.

## Round 2 — review
Reviewers: MNT (MNT-1, MNT-2 were Fix-now majors). Skipped: CON (rated CON-2 minor; the shared finding goes to MNT), SEC+PRF (no findings), UX (the `macos/App/**` change is two doc comments).

MNT-1 **Resolved**, MNT-2 **Resolved**. No new blocker, so no Fix plan 2. Below the round-2 bar, for the user to place: `engine.proto:1291` with its generated copies and `GenealogyStore.swift:1515` still say Persons order by the title shown; `docs/ideas/international-names.md:37` names the now-private `nameformats.Format`.

**User answer (2026-11-06).** All three follow-ups go on this PR. Not review findings; no review round follows, by the two-round cap.

### Fix plan 2
**F-8 — Persons list comments say the sort title** (follow-up 1)
Files: `api/proto/engine.proto`, its generated copies, `macos/App/Platform/GenealogyStore.swift`.
Do: `engine.proto:1291`: Persons order by the name without honorifics (auto-reconciled name, else label), then ref-only rows. Regenerate the copies with the repo's generator, not by hand. `GenealogyStore.swift:1515`: "by sort name, which leaves honorifics out, then ref-only by ref". Comments only.
Check: the diff of the Swift and generated files changes only comment lines.

**F-9 — Idea file names `Profile.Format`** (follow-up 2)
Files: `docs/ideas/international-names.md`.
Do: line 37: `nameformats.Format` → `nameformats.Profile.Format`.
Check: `grep -rn "nameformats\.Format" docs/ideas` prints nothing.

**F-10 — `SourcePageModelTests` gets its own folder per test** (follow-up 3)
Root cause: the suite's fixed `projectDir` is `/tmp/source-page.provenencia`, tests run in parallel, and one test's `defer removeItem` deleted the folder under another. The failure read "That file is missing from the project folder."
Files: `macos/ProvenenciaTests/SourcePageModelTests.swift`.
Do: `private let projectDir = ArtifactFileFixtures.uniqueProjectDir()`. Swift Testing makes a new suite value per test, so each gets its own. Leave the `defer`s. Make the same change in any other suite whose fixed `/tmp` folder is removed in a test; say which.
Check: `scripts/check-pr.sh` PASS.

## Round 2 — fix
Head `7d8694f8`. Gate PASS (orchestrator run). F-8, F-9, F-10 done; the Swift and generated diffs are comment lines plus the two test `projectDir` lines. F-10 also moved `EvidenceGraphModelTests`, whose fixed folder a test removed. `WorkspaceDestinationHostTests` stays: its removed paths are per-test folders. Not reviewed: no third round.

## Result
Gate PASS at `7d8694f8`; `origin/main` has not moved since the branch was cut · CI green at `c4ea0c09`, not yet read for the follow-up commits · open: none · follow-ups: none, all three done on this PR · to user: answered (MNT-1, MNT-5, MNT-7).
