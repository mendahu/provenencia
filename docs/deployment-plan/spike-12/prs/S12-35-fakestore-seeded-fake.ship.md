# Ship log — S12-35 FakeStore is a seeded fake

**Plan:** [`S12-35-fakestore-seeded-fake.md`](S12-35-fakestore-seeded-fake.md) · **Branch:** `s12-35-fakestore-seeded-fake` · **PR:** [#617](https://github.com/mendahu/provenencia/pull/617)
**State:** ready

## Round 0 — implement
Head `5667b777`. Gate PASS (orchestrator re-ran `scripts/check-pr.sh`). Report: none deviations.

## Round 1 — review
Reviewers: CON, MNT, UX (`macos/App/Platform/FakeStore.swift`), SEC+PRF (`api/ffi/swift_vocabulary_test.go`). Skipped: —.

| ID | Lens | Severity | Title | Decision | Evidence / reason |
| --- | --- | --- | --- | --- | --- |
| MNT-1 | MNT | major | Delete dead `unique` helper | Fix now → F-1 | verified: only definition; no call sites |
| CON-1 | CON | minor | `copySourceError` has no test | Fix now → F-2 | verified: hook present; no test references |
| MNT-2 | MNT | minor | S12-35 map Ships omits `nextWriteEffects` | Fix now → F-3 | verified: Ships vs standing rule |

Nits (logged): `CON-2 · CitationComposerModelTests.swift:541 · Rename stepperListsOnlyThisGraphsCitations` · `CON-3 · SourcePageModelTests.swift:1203 · Rename graphRowsAreTheSourcesGraphsByName` · `MNT-3 · FakeStore.swift:3375 · Comment that listCitationsByArtifact does not filter by graph membership`

### Fix plan 1
**F-1 — Remove dead `unique` helper** (from MNT-1)
**F-2 — Test `copySourceError`** (from CON-1)
**F-3 — Name `nextWriteEffects` in S12-35 Ships** (from MNT-2)

## Round 1 — fix
Head `fef49d45`. Gate PASS (orchestrator). F-1 done; F-2 done (`copySourceErrorIsScripted`); F-3 done. CON-1 / MNT-2 Resolved by orchestrator read.

## Round 2 — review
Reviewers: MNT only (had Fix-now major). Scope: `5667b777...fef49d45`.

| ID | Lens | Severity | Title | Decision | Evidence / reason |
| --- | --- | --- | --- | --- | --- |
| MNT-1 | MNT | major | Delete dead `unique` | Resolved | `rg` empty for `unique` in FakeStore |
| — | MNT | — | new blockers | — | none |

Nits (logged): `MNT-4 · FakeStore.swift:3536 · Reword insertSource doc away from “like the engine”`

## Result
Gate PASS · CI pending · open: none · follow-ups: none · to user: — · CON miss: none
