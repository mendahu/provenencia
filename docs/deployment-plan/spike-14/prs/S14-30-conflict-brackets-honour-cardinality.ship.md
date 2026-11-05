# Ship log — S14-30 Conflict brackets honour cardinality

**Plan:** [`S14-30-conflict-brackets-honour-cardinality.md`](S14-30-conflict-brackets-honour-cardinality.md) · **Branch:** `s14-30-conflict-brackets-honour-cardinality` · **PR:** #713
**State:** ready

## Round 0 — implement
Head `7e69154a`. Gate PASS (orchestrator re-ran `scripts/check-pr.sh`). Report: plan Context / map dogfood rows use plain text for the deleted UX note; manual check §10 not run; version 0.3.2 (not 0.3.1 planning base).

## Round 1 — review
Reviewers: CON (Composer), MNT (Grok 4.7), UX (Grok 4.7), SEC+PRF (Grok 4.7).
Skipped: none — UX matches `macos/App/**`; SEC+PRF matches `macos/App/Features/**` models (`PropertiesModel`, Evidence graph card paint).
Base `3a48302716575f15095a9d72ade26ab13d78042e`. Head `7e69154a905adb4a4f7e80630371a1afafa6dfa6`.

| ID | Lens | Severity | Title | Decision | Evidence / reason |
| --- | --- | --- | --- | --- | --- |
| CON-1 | CON | minor | Plan §10 manual check not performed | Rejected | ship-pr §8 Finish assigns §10 to the user after ready-for-review; T-1–T-7 are implementer AC evidence; §10 stays on the user report / Manual check |
| CON-2 | CON | minor | No test for `PropertiesSnapshot.cardinalityByKey` | Fix now → F-1 | verified: `PropertiesSnapshot.swift:47-56` implements D-4; `rg cardinalityByKey macos/ProvenenciaTests` only hits marks tests with explicit maps |

Nits (logged): `CON-3 · S14-30 plan Landed note · could list §10 under deviations (optional; already inline)`

MNT: Clean · UX: Clean · SEC+PRF: Clean

### Fix plan 1

**F-1 — Test `PropertiesSnapshot.cardinalityByKey` D-4 edges** (from CON-2)
Files: `macos/ProvenenciaTests/PropertiesModelTests.swift` (suite already builds `PropertiesSnapshot`). Skill: `add-swift-test`.
Failing test first: `cardinalityByKeyNormalizesEmptyAndPrefersMultiple` — assert empty string → `.single`; duplicate key with `.single` + `.multiple` → `.multiple`.
Do: Add one or two `@Test`s that construct `PropertiesSnapshot` with `CatalogProperty` rows (`cardinality: ""`, and two rows same key / different origin with single vs multiple) and assert `cardinalityByKey` matches D-4. Reuse existing property helpers when they fit.
Check: new tests pass; `rg -n 'cardinalityByKey' macos/ProvenenciaTests` shows the snapshot suite; `scripts/check-pr.sh` PASS.

## Round 1 — fix
Head `d31e7b29`. Gate PASS (orchestrator re-ran). F-1 done: `PropertiesModelTests/cardinalityByKeyNormalizesEmptyAndPrefersMultiple` covers empty→`.single` and duplicate-key prefer `.multiple`. CON-2 Resolved (orchestrator read the test). No round 2 (no Fix-now blocker or major).

## Result
Gate PASS · CI (pending ready) · open: none · follow-ups: none · to user: — · CON miss: none
