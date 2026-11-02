# Ship log — S14-21 Search finds every spelling

**Plan:** [`S14-21-search-finds-every-spelling.md`](S14-21-search-finds-every-spelling.md) · **Branch:** `s14-21-search-finds-every-spelling` · **PR:** #684
**State:** ready

## Round 0 — implement
Head `7a921743`. Gate PASS (orchestrator run). PR #684 (draft).
Report deviations: (1) T-4 queries "Søby-Kirke" not "Soby-Kirke" — under D-4 a no-fold-letter hyphenated query goes out as one phrase that cannot span title and `folded` columns; (2) `graphcache.RowOf(v, property)` exported, takes Property id; (3) new exported `searchindex.QueryPart`; (4) `core/search/graphs_test.go` compares to `searchindex.ProjectionVersion` (now 10); (5) Person `alias:` rows in separate `TestMatchTextCarriesFoldedInNicknames`; (6) stale `rawTokens` comment fixed in `TextMatch.swift`; (7) manual check not run.
STEP-12 (landed-docs commit) not done: the implementer's commit was denied by the auto-mode classifier. Taken to the user.

## Round 1 — review
Reviewers: CON (Sonnet), MNT, UX, SEC+PRF. Skipped: none (UX: `macos/App/Platform/**`; SEC+PRF: migration 000060, `core/valuecompare`, `namealiases` tests, `core/database/**`, `core/search`).

| ID | Lens | Severity | Title | Decision | Evidence / reason |
| --- | --- | --- | --- | --- | --- |
| MNT-1 (+ SEC/PRF outside-lens note) | MNT | blocker → major | `autoreconciler.sameDisplay` copies `graphcache.sameValues`' per-property row walk; `RowOf` exported to feed it | Fix now → F-1 | verified: both walk property count, per-property length, pairwise compare; only the row test differs. D-11 planned `rowOf` unexported. Same-pass cleanup of code this PR adds |
| UX-1 | UX | major | Tag word `alias` is searchable: "Al", "Ali", "Alias" return every Person whose shown name has nicknames | To user → Fix now → F-2 | verified in a scratch copy of head: lone "Thomas Smith" is returned for "Al", "Ali", "Alias" (reason `other`). "Nam"/"Topo" give 0 here only because the doc has no alternates. The fix changes D-8's tag strings. User 2026-11-02: letterless tags for all three (name, toponym, alias) |
| CON-1 | CON | major | STEP-12 / DoD docs not on the branch | To user → Fix now → F-5 | implementer's commit denied by the auto-mode classifier; the orchestrator does not re-run a denied action. User 2026-11-02: approved the STEP-12 docs commit in the fix round |
| PRF-1 | PRF | major | Joining a new spelling onto the 6,000-event city still reprojects all dependents (~2.1 s) | Follow-up | measured by reviewer at Scale: new spelling 2.11 s vs no-move join 97 ms; cost predates this PR and is not made worse; D-9 rejects the warm-graph route. Note: the plan's Scale calls this case a "rename"; adding a spelling from a new register is a normal Promote action. Idea file if the user agrees |
| CON-2 | CON | minor | "Soby-Kirke" (hyphen, no ø) misses "Søby Kirke register"; T-4 queries "Søby-Kirke" | Follow-up | behavior follows D-4 as written; T-4's example contradicts D-4. Recorded in the Landed note; reopening D-4's split is a later change |
| MNT-2 (+ CON-4) | MNT | minor | `valuerules.Product()` built twice (searchindex `productRules`, autoreconciler `ValueRules`), not memoized | Fix now → F-2 | verified: `rg 'valuerules.Product()'` → two package vars. Minor in files this PR changes |
| MNT-3 | MNT | minor | New `f.place` helper not used by `locatedBirth` and the inline place subjects in `handles_test.go` | Fix now → F-3 | same-pass cleanup; helper and copies both in this PR's test file |

Nits (logged): `MNT-4 · searchindex.go QueryPart · accent-only key computed twice on the no-fold path` · `MNT-5 · QueryPart is new and exported, not in Codebase fit; name it in the Landed note` · `CON-3 · searchindex.go midWordApostrophe · checks IsLetter/IsDigit, plan says "kept rune"; align comment or check`

### Fix plan 1

**F-1 — One per-property row walk in `graphcache`; `rowOf` unexported again** (from MNT-1)
Files: `core/database/graphcache/verify.go`, `core/database/autoreconciler/autoreconciler.go`. Skill: `add-match-rule` is not needed.
Failing test first: none (refactor; `TestJoinReprojectsDependentsOnlyWhenTheRootMoves`, the graphcache Verify tests and `PROVENENCIA_GRAPH_VERIFY=1` promotealign tests must stay green).
Do: Split `sameValues(got map[string][]Value, want []TruthRow)` into one unexported walk that takes the per-row test (`func(Value, TruthRow) bool`): bucket `want` by `PropertyID`, sort each bucket by `Rank`, compare the property count, each property's row count, then each pair with the test. Verify keeps calling it with `valueEqual` (and keeps its error text). Add one exported entry point `SameDisplayRows(stored map[string][]Value, written []TruthRow) bool` that runs the same walk with `SameDisplay(rowOf(v), t)`. Rename `RowOf` back to unexported `rowOf` (D-11). In `autoreconciler`, `recomputeBatch` collects a flat `[]TruthRow` per entity and calls `graphcache.SameDisplayRows`; delete `autoreconciler.sameDisplay`. An id with no stored rows and no written rows stays unmoved.
Check: `rg -n "graphcache\.RowOf|func RowOf|func sameDisplay\b" core` prints nothing; `rg -n "sort\.Slice\(" core/database/graphcache/verify.go` prints exactly one line (the walk); scripts/check-pr.sh PASS.

**F-2 — Secondary tags carry no letters** (from UX-1; supersedes D-8's tag strings, user 2026-11-02)
Files: `core/database/searchindex/searchindex.go`, `core/search/handles_test.go`, `.agents/skills/add-searchable-kind/SKILL.md` and `docs/conclusion-reconciliation.md` if either quotes the tag strings.
Failing test first: `TestNicknameFindsTheFormalName/a tag word finds nothing` — promote a lone `f.person("Thomas Smith")`; `f.search(q, KindPerson)` for each of "Al", "Ali", "Alias" returns no hits. Add a second subtest `TestNicknameFindsTheFormalName/an alternate's tag word finds nothing` — promote `f.person("Tom Smith", "Thomas Smith")` and a place `f.place("York", "Yorke")`; "Nam" (KindPerson) and "Topo" (KindPlace) return no hits. Run red first at the current head.
Do: Change the three constants to tags made only of ASCII punctuation and a tab, which unicode61 and `valuecompare` both drop as separators: `SecondaryTagName = "=\t"`, `SecondaryTagToponym = "@\t"`, `SecondaryTagAlias = "~\t"`. Nothing else changes: `alternateMatch` already parses by constant, and `ProjectionVersion` is already bumped on this branch (no second bump). Update any doc or guidance that quotes the old strings.
Check: both subtests pass; `TestNicknameFindsTheFormalName`, `TestMatchTextListsEachAlternateOnce`, `TestMatchTextCarriesFoldedInNicknames` pass; `rg -n '"(name|toponym|alias):\\t' core .agents docs/*.md` prints nothing; scripts/check-pr.sh PASS.

**F-3 — One product rules value** (from MNT-2, CON-4)
Files: `core/database/valuerules/*.go`, `core/database/searchindex/handles.go`, `core/database/autoreconciler/autoreconciler.go`.
Do: Make `valuerules.Product()` build once and return the same value on every call (`sync.OnceValue`). Keep both package vars as they are. If any caller mutates the returned rules (a test that assigns into its maps or fields), stop and report it instead of memoizing.
Check: `rg -n "valuerules\.Product\(\)" core` prints the two package vars and the tests only; scripts/check-pr.sh PASS.

**F-4 — `locatedBirth` and the inline place subjects use `f.place`** (from MNT-3)
Files: `core/search/handles_test.go`.
Do: Move `(f *handleFixture) place` next to `person`. `locatedBirth` and the inline place subjects (about lines 352, 367, 393, 478 at `7a921743`) call `f.place(...)`.
Check: `rg -n 'PropertyID: f.prop\("toponym"\).ID' core/search/handles_test.go` prints only the line inside `place`; `go test -tags fts5 ./core/search/` PASS.

**F-5 — STEP-12 docs** (from CON-1; the user approved this commit 2026-11-02)
Files: the plan, `deployment-plan.md`, `completed.md`, as STEP-12 says.
Do: STEP-12 as written, PR #684, date 2026-11-02, in one commit "Record S14-21 as landed in pull request 684.". The Landed note lists: the seven round-0 deviations (ship log Round 0); `searchindex.QueryPart` is new and exported (MNT-5); CON-2 — under D-4, "Soby-Kirke" (hyphen, no ø) does not find "Søby Kirke register", so T-4 queries "Søby-Kirke" (follow-up); D-8's tag strings changed to `=\t`, `@\t`, `~\t` (UX-1, user 2026-11-02); `rowOf` stays unexported and the row walk is shared (F-1); PRF-1 follow-up (a join that adds a new spelling to the 6,000-event city still reprojects every dependent, 2.1 s).
Check: scripts/check-pr.sh PASS (plan-docs).

## Round 1 — fix
Head `a17a1589`. Gate PASS (orchestrator run). F-1 `ca69f77a`; F-2 + F-4 `06c093c1` (one commit, both in `handles_test.go`; F-2 run red first); F-3 `5102398d`; F-5 `a17a1589`. Report deviations: map row uses `Landed #684` (repo row form); Landed note also records the memoized `Product()`; manual check not run.
Minors checked by the orchestrator: MNT-2 **Resolved** (`valuerules.Product` is `sync.OnceValue`, both package vars read it); MNT-3 **Resolved** (`rg 'PropertyID: f.prop("toponym").ID'` prints only the line in `place`).

## Round 2 — review
Reviewers: CON (Sonnet; CON-1), MNT (MNT-1), UX (UX-1). Not sent: SEC+PRF (PRF-1 was Follow-up, no Fix-now blocker or major). Scope `7a921743...a17a1589`.

| ID | Lens | Severity | Title | Decision | Evidence / reason |
| --- | --- | --- | --- | --- | --- |
| CON-1 | CON | major | STEP-12 docs | Resolved | `a17a1589`; check-plan-docs 0 errors |
| MNT-1 | MNT | major | Shared row walk | Resolved | one `walkValues` in `verify.go`; `rowOf` unexported; `autoreconciler.sameDisplay` deleted |
| UX-1 | UX | major | Tag word searchable | Resolved | tags `=\t`, `@\t`, `~\t`; new negative subtests pass on head |

No new blockers. Nits (logged): `MNT-6 · autoreconciler.go:348 · recomputeBatch comment names SameDisplay, now calls SameDisplayRows` · `UX-2 · registry.go:290 · comment still says "(name: / toponym:)"` · `UX outside-lens · plan D-8 and the handles.go Design row quote the old tag strings; the Landed note records the change`

## Result
Gate PASS (head `a17a1589`, origin/main unmoved at `25ed4bd6`) · CI pending · open: none · follow-ups: PRF-1 (new-spelling join onto the 6,000-event city reprojects every dependent, 2.1 s; needs a call against D-9), CON-2 ("Soby-Kirke" with hyphen and no ø misses "Søby Kirke register" under D-4) · to user: UX-1 (letterless tags, all three), CON-1 (STEP-12 commit approved) — both answered · CON miss: none
