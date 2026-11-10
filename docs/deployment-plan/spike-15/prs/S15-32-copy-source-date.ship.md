# Ship log — S15-32 Copy Source can carry the date

**Plan:** [`S15-32-copy-source-date.md`](S15-32-copy-source-date.md) · **Branch:** `s15-32-copy-source-date` · **PR:** #778
**State:** ready

## Round 0 — implement
Head `f6474543`. Gate PASS (orchestrator run, against `cd035824`). PR #778 (draft). Report deviations: `main`'s `VERSION` was 0.7.1, so the bump is 0.7.1 → 0.8.0 per the header's rule; the dangling-date `TestCopy` case first deadlocked on a held connection, fixed in `853a9e23`; the branch was pushed once before the first full gate run; the manual check and board comparison were not run. The last commit is unpushed: DNS for github.com fails on this Mac.

## Round 1 — review
Reviewers: CON, MNT, UX, SEC+PRF. Skipped: none. UX for `macos/App/**` and `Localizable.xcstrings`; SEC+PRF for `api/ffi/**` and because `sourcecopy` reads the original's stored `date_values` row and copies it.

| ID | Lens | Severity | Title | Decision | Evidence / reason |
| --- | --- | --- | --- | --- | --- |
| MNT-1 | MNT | major | `revisionChanges` is a fourth hand-written latest-revision reader | Fix now → F-1 | Same-pass cleanup: this PR adds a copy. Confirmed by body grep `rg -n "FROM audit_changes" -g '*_test.go'`: latest-revision readers in sourcecopy, promote (×2), deleteimpact, canonicalentities, identityclaims, artifacts, citations, subjects (×2) |
| MNT-2 | MNT | minor | `dateRows` copies the `sources` test helper | Fix now → F-2 | Same-pass cleanup: `sources_test.go:971` matches line for line; named counters also in observations, deleteimpact, ingest |
| CON-1 | CON | major | Manual check and board comparison not run | To user (manual check at finish) | Not a code defect: the ship-pr finish step hands the plan's §10 manual check to the user before merge. Listed in the report and PR body |
| SEC-1 | SEC | minor | Date phrase has no length bound; the copy now duplicates it | Follow-up | `datevalue.Validate` is unchanged by this PR; a crafted-catalog-only cost (one high cap, per the caps rule). The plan's Edges line "bounds its fields" overstates this; noted here, plan not edited |

UX: no findings. PRF: no findings.

Nits (logged): `CON-2 · core/database/sourcecopy/sourcecopy.go:100-101 · reflow the package doc` · `MNT-3 · core/database/sourcecopy/sourcecopy.go:4 · same reflow (deduped with CON-2; rides along in F-3)` · `MNT-4 · core/database/sourcecopy/sourcecopy_test.go:700-724 · FK-off dangling setup is a second copy of sources_test.go dangle; share on a third`

### Fix plan 1

**F-1 — One shared latest-revision reader in `dbtest`** (from MNT-1)
Files: `core/database/dbtest/audit.go` (or a new `dbtest/revision.go`), and the test files below.
Failing test first: none needed (test-helper refactor); every moved test must still pass.
Do: add `dbtest.LatestRevision(t testing.TB, db *sql.DB) (action string, changes []rowchange.Change)`: the newest `audit_transactions` row's `action_type`, and its `audit_changes` (`entity_type`, `entity_id`, `action`, `changes_json` decoded into `Fields`) in write order (`ORDER BY ch.rowid`). Taking `*sql.DB` keeps `dbtest` free of `core/database`. Move onto it: `sourcecopy_test.go` `revisionChanges` (delete it); `promote/promote_test.go` `lastRevisionChanges` (line 229) and the inline query at line 58 (sort by entity type in Go where the test needs that order); `deleteimpact/conclusion_test.go` `lastRevision` (line 165); the inline queries at `canonicalentities_test.go:105`, `identityclaims_test.go:188`, `artifacts_test.go:119`, `citations_test.go:244`, `subjects/subjects_test.go:95` and `:231` (the last change of the latest revision is the last element). Leave `core/database/audit/*_test.go` (they test the audit writer itself and read by transaction id) and `core/writes/run_test.go` (reads a given revision number).
Check: `rg -n "FROM audit_changes" --type go -g '*_test.go'` prints only lines in `core/database/audit/` and `core/writes/run_test.go`, plus the new helper in `core/database/dbtest/`; `rg -n "func (revisionChanges|lastRevisionChanges|lastRevision)\b" --type go` prints nothing; scripts/check-pr.sh PASS.

**F-2 — One shared row counter in `dbtest`** (from MNT-2)
Files: `core/database/dbtest/`, `sourcecopy_test.go`, `sources/sources_test.go`, `observations/observations_test.go`, `deleteimpact/release_test.go`, `core/ingest/stage_test.go`.
Failing test first: none needed (test-helper refactor).
Do: add `dbtest.Count(t testing.TB, q interface{ QueryRow(string, ...any) *sql.Row }, query string, args ...any) int` (works for `*sql.DB` and `*sql.Tx`). Replace both `dateRows` helpers (delete them; call `dbtest.Count(t, db, "SELECT COUNT(*) FROM date_values")`), `observations` `rowCount`, `deleteimpact` `countRows`, and `ingest` `fileRows` with it. Leave `api/ffi/handlers` `countCatalogRows` (opens the catalog by directory) but make its body call `dbtest.Count`. Inline one-off `SELECT COUNT(*)` statements elsewhere are not helper copies; leave them.
Check: `rg -nE "^func \w+\(t \*testing\.T, (c \*database\.Catalog|tx \*sql\.Tx)(, query string, args \.\.\.any)?\) int \{" --type go` prints nothing; scripts/check-pr.sh PASS.

**F-3 — Reflow the `sourcecopy` package doc** (nits CON-2 / MNT-3)
Files: `core/database/sourcecopy/sourcecopy.go`.
Do: reflow the package doc paragraph so "It" no longer sits alone on a short line.
Check: gofmt clean; scripts/check-pr.sh PASS.

## Round 1 — fix
Head `3069e889` (commits `3555248a`, `3069e889`); branch fully pushed. F-1 done: `dbtest.LatestRevision`, every latest-revision reader outside `core/database/audit/` and `core/writes/run_test.go` moved onto it. F-2 done: `dbtest.Count` replaces both `dateRows`, `rowCount`, `countRows`, `fileRows`, and (found by the Check's grep) `auditCount`, `searchCount`, `auditTotal`; `countCatalogRows` calls it. One Check hit remains by design: `observations_test.go` `objectCount` walks `objects/` on disk, not a row count. Added beyond the plan: `dbtest.MustDB`, `dbtest.EntityTypes`. F-3 done (nits CON-2 / MNT-3): orchestrator read `3555248a`, **Resolved**.

## Round 2 — review
Reviewers: MNT (MNT-1 was a Fix-now major). Not sent: CON (CON-1 is a To-user manual check, not a Fix-now), UX and SEC+PRF (no Fix-now findings). MNT-2 and the nits ride along with MNT.

| ID | Lens | Severity | Title | Decision | Evidence / reason |
| --- | --- | --- | --- | --- | --- |
| MNT-1 | MNT | major | (round 1) fourth latest-revision reader | Resolved | `dbtest.LatestRevision`; `rg "FROM audit_changes" -g '*_test.go'` prints only `core/database/audit/*`, `core/writes/run_test.go` (by revision number) and `subjects/delete_test.go:189` (a filtered count) |
| MNT-2 | MNT | minor | (round 1) `dateRows` copy | Resolved | orchestrator read `3069e889`: `dbtest.Count` replaces every named counter except `objectCount` (walks `objects/` on disk) |

No new blockers.

Nits (logged): `MNT-5 · core/database/dbtest/revision.go:66 · dbtest.MustDB leaves private copies in graphcache/inputs_test.go:162 and namealiases/namealiases_test.go:209 (14 call sites)` · `MNT-6 · sourcecopy_test.go:605, deleteimpact/conclusion_test.go:139, deleteimpact/facets_test.go:96, promote/promote_test.go:128 · still spell out c.DB() + t.Fatal` · `MNT-7 · core/writes/run_test.go:50, core/derivatives/thumbnail_test.go:79 · dbtest.Count(t, dbtest.MustDB(t, c), "…audit_transactions") repeated; a local or dbtest.Revisions helper` · `MNT-8 · core/database/dbtest/revision.go · Count and MustDB are not about revisions; move to dbtest.go`

MNT-5 has the same-pass cleanup shape, but round 2 can only add blockers and a Fix plan 2 exists only for round-2 blockers, so it goes to the user rather than into a fix round.

Round-2 head `3069e889`: gate PASS (orchestrator run). `git fetch origin` fails (DNS), so whether `main` moved since `cd035824` is unchecked.

### Fix plan 2
The user asked to fold the round-2 nits MNT-5 to MNT-8 into this PR (2026-11-10). Test-only; not a review round. The orchestrator reads the commits and runs the gate.

**F-4 — One `MustDB`** (from MNT-5, MNT-6)
Files: `core/database/graphcache/inputs_test.go`, `core/database/namealiases/namealiases_test.go`, `core/database/sourcecopy/sourcecopy_test.go`, `core/database/deleteimpact/conclusion_test.go`, `core/database/deleteimpact/facets_test.go`, `core/database/promote/promote_test.go`.
Do: delete the private `mustDB` helpers at `graphcache/inputs_test.go:162` and `namealiases/namealiases_test.go:209`; point their call sites at `dbtest.MustDB`. Replace the hand-written `db, err := c.DB(); if err != nil { t.Fatal(err) }` at the four MNT-6 sites with `db := dbtest.MustDB(t, c)`. Find others by body in the files this PR already touches only.
Check: `rg -n "^func mustDB" --type go` prints nothing; scripts/check-pr.sh PASS.

**F-5 — Name the revision count** (from MNT-7)
Files: `core/database/dbtest/`, `core/writes/run_test.go`, `core/derivatives/thumbnail_test.go`, and any other file this PR moved onto `dbtest.Count(... "SELECT COUNT(*) FROM audit_transactions")`.
Do: add `dbtest.Revisions(t testing.TB, db *sql.DB) int` (the `audit_transactions` count) and use it at those sites.
Check: `rg -n 'dbtest.Count\(.*FROM audit_transactions' --type go` prints nothing; scripts/check-pr.sh PASS.

**F-6 — File placement** (from MNT-8)
Files: `core/database/dbtest/revision.go`, `core/database/dbtest/dbtest.go`.
Do: keep `LatestRevision`, `EntityTypes` and `Revisions` in `revision.go`; move `Count` and `MustDB` to `dbtest.go`.
Check: scripts/check-pr.sh PASS.

## Round 2 — fix
Head `7cb08479` (commits `aa9de12e`, `7cb08479`). Gate PASS (orchestrator run). F-4: `rg -n "^func mustDB" --type go` prints nothing; the private copies in graphcache and namealiases, the MNT-6 sites, and the same pattern in the other test files this branch touches use `dbtest.MustDB`. F-5: `dbtest.Revisions`; `rg -n 'dbtest.Count\(.*FROM audit_transactions'` prints nothing. F-6: `Count` and `MustDB` in `dbtest.go`. MNT-5 to MNT-8 **Resolved** (orchestrator read the commits).

## Result
Gate PASS on `7cb08479` · `origin/main` still `cd035824`, no merge needed · CI green (localizable, test ×2) · open: none · follow-ups: SEC-1 (home to be chosen by the user) · to user: CON-1 manual check and board comparison
