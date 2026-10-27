# Ship log — S11-44 Des as a nickname of Desiree and Desirae

**Plan:** [`S11-44-des-nickname.md`](S11-44-des-nickname.md) · **Branch:** `s11-44-des-nickname` · **PR:** #594
**State:** ready

## Round 0 — implement
Head `1d383937`. Gate PASS (orchestrator run against `e7411828`). Report: no deviations; manual check §10 not run.

## Round 1 — review
Reviewers: CON (Sonnet), MNT, SEC+PRF. Skipped: UX (no `macos/App/**`, strings, or error codes). SEC+PRF routed by `core/database/namealiases` (matching layers) and `core/database/**`.

| ID | Lens | Severity | Title | Decision | Evidence / reason |
| --- | --- | --- | --- | --- | --- |
| MNT-1 | MNT | minor (rated major) | Seed header, NOTICE and `seeded-vocabulary.md` §4.5 still say every set comes from names.csv | Fix now → F-1 | verified: `nicknames_seed.go:1-23` "Changed from that CSV" lists merges, removals and tokens only; NOTICE lines 4–11 say the same; §4.5 "The list is names.csv". The two Des sets are the first product-authored rows. Re-rated minor: provenance text, no behavior; Fix now because this PR changes the file |

Nits (logged): `CON-1 · namealiases_test.go:178 · not table-driven (plan's T-1 shape)` · `MNT-2 · namealiases_test.go:178 · one test function per PR's sets; a table of product-seed pairs would make the next one a row` · `MNT-3 · namealiases_test.go:194 · say in a comment that no comparison counts as below credit`

Outside lens (MNT): `namealiases.AddProductNicknames` (`core/database/namealiases/add.go:26`) already copies new product sets into an existing catalog; it is what the user's one-off command can call. Not a finding (D-3); passed to the user.

SEC+PRF: clean. CON: clean.

### Fix plan 1
**F-1 — Say the seed has product-added sets** (from MNT-1)
Files: `core/database/namealiases/nicknames_seed.go` (header comment), `core/database/namealiases/third_party/nicknames/NOTICE`, `docs/seeded-vocabulary.md` §4.5.
Failing test first: none — comment and doc text only.
Do: in the header's "Changed from that CSV" paragraph, add one sentence: sets not in names.csv were added after review — desirae/des and desiree/des (S11-44, 2026-10). Add the matching sentence to the NOTICE's "This file was changed from names.csv" paragraph, and mention there that pairs joining two different given names were removed (as the header already says). In `docs/seeded-vocabulary.md` §4.5 change "The list is names.csv from …" to say the list starts from names.csv, with reviewed removals and additions. Do not change any code.
Check: scripts/check-pr.sh PASS.

## Round 1 — fix
Head `04627ccd`. Gate PASS (orchestrator run against `e7411828`). F-1 done: MNT-1 **Resolved** (orchestrator read the commit: header, NOTICE and §4.5 now name the product-added sets). No round 2: no Fix-now blocker or major.

## Result
Gate PASS · CI pending · open: none · follow-ups: none · to user: — · CON miss: none
