# Ship log — S11-48 Religion and languages spoken on a person

**Plan:** [`S11-48-religion-languages.md`](S11-48-religion-languages.md) · **Branch:** `s11-48-religion-languages` · **PR:** —
**State:** implementing

The branch already carries the plan, its review log and `docs/ideas/organizations.md` (commits `5badc2ad`, `f53cc97a`) with `origin/main` @ `846fa656` merged in (`fd21f78e`). `main` moved past the planning base only in S11-41's files; no Current state row is touched. The plan and the code ship as one PR.

## Round 0 — implement
Head `1fa2bd10`. Gate PASS (orchestrator's run, against `846fa656`). Report: one deviation — `TestVocabularyDropsOnBindingReorder` also hard-coded the person order and now reads `seededPersonOrder()` (plan risk row 1; in the Landed note). The implementer's own gate hit two timing flakes outside this diff (`promotealign` `TestLargeSourceBudgets`; `OmnibarResultsModelTests` / `CatalogSearchTests`), which passed on rerun. Manual check not run.

## Round 1 — review
Base `846fa656`, head `1fa2bd10`. Reviewers: CON (Sonnet), MNT, SEC+PRF (`api/ffi/handlers/subject_defs_test.go`, `core/database/**`, `core/autoreconcile` test). Skipped: UX (no `macos/App/**`, strings, or error codes).

| ID | Lens | Severity | Title | Decision | Evidence / reason |
| --- | --- | --- | --- | --- | --- |
| MNT-1 | MNT | major | `TestInstallSeedsLikeness` and `TestInstallSeedsFuneralAndPallbearer` are now dead copies of `TestInstallMatchesRegistry`; two pin shapes | Fix now → F-1 | verified: both tests only round-trip registry values (`subjectvocab_test.go` `TestInstallSeedsLikeness`, `TestInstallSeedsFuneralAndPallbearer`); `Neutral` is already covered by `TestInstallFlagsNeutralTerms`. Same-pass cleanup: this PR adds the generic test the copies duplicate, so it supersedes D-7's "the three older Install tests stay"; no researcher-visible change |
| MNT-2 | MNT | major | `add-seeded-vocabulary` does not teach the registry round-trip pattern | Fix now → F-2 | verified: the subject-vocabulary section says only "tests that create installs and open does not heal" |

Nits (logged): `CON-1 · ship log uncommitted (the orchestrator commits it)` · `MNT-3 · subjectvocab_test.go · local sort := -1 reads as the package; name it order (F-1 rewrites it)` · `MNT-4 · subjectvocab_test.go:581 · report a missing binding apart from a mismatch` · `MNT-5 · graphcache/vocab_test.go:110 · slices.SortFunc instead of a new sort import`

CON: clean. SEC+PRF: clean.

### Fix plan 1

**F-1 — One pin shape for product seed values** (from MNT-1, MNT-3)
Files: `core/database/subjectvocab/subjectvocab_test.go`. Skill: `add-seeded-vocabulary`.
Failing test first: none — a test refactor; every assertion it removes is still made by `TestInstallMatchesRegistry` or by the new table.
Do:
- Replace `TestSeedReligionAndLanguagesSpoken` with one table-driven `TestSeedProductValues` that reads the registry (`seedProperties`, `seedTerms`, `BindingsForType`, `LockedBinding`) and pins product values as literals. Cases: `religion` and `languages_spoken` (everything `TestSeedReligionAndLanguagesSpoken` checks today); `likeness` (depiction, multiple, "Likeness", locked, person sort 1); `gender` (term, single, unlocked, sort 3, terms `man`, `woman`, `non_binary` none neutral); `interests` (text, multiple, unlocked, sort 4); `event_type` terms `engagement` "Engagement", `travel` "Travel", `funeral` "Funeral"; `role` term `pallbearer` "Pallbearer". Take every literal from the deleted tests as they are today. The local sort-order variable is named `order`, not `sort`.
- Delete `TestInstallSeedsLikeness` and `TestInstallSeedsFuneralAndPallbearer`.
- Cut `TestInstallSeedsEngagementTravelInterestsAndGender` to the user-term create it alone covers (the `users.Upsert` + `writes.Run` `create_property_term` "Agender" block and the lookups it needs), renamed `TestInstallLetsAUserTermJoinASeededProperty`.
- Plan Landed note: one line saying F-1 supersedes D-7's "the three older Install tests stay" (same-pass cleanup, review MNT-1).
Check: `grep -n "func TestInstallSeeds" core/database/subjectvocab/subjectvocab_test.go` prints nothing; `CGO_ENABLED=1 go test -tags fts5 ./core/database/subjectvocab/...` passes; scripts/check-pr.sh PASS.

**F-2 — The skill teaches the pattern** (from MNT-2)
Files: `.agents/skills/add-seeded-vocabulary/SKILL.md`.
Do: in "Subject vocabulary / credibility grades", after "tests that create installs and open does not heal.", add: "For `subjectvocab`, `TestInstallMatchesRegistry` already checks every seeded Property, term and binding against the registry, so a new row needs no Install test: add its product values (labels, value type, cardinality, sort order, lock) as a case in `TestSeedProductValues`. Tests in other packages read the registry (`BindingsForType`, `len(seedProperties)`) instead of restating seed lists or counts. `core/claimevidence`'s `seedOrder` is a matching fixture, not a copy of the seed, and stays as it is."
Check: scripts/check-pr.sh PASS.
