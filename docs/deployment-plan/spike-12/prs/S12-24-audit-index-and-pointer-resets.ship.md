# Ship log — S12-24 Audit history: indexed note authors, audited pointer resets

**Plan:** [`S12-24-audit-index-and-pointer-resets.md`](S12-24-audit-index-and-pointer-resets.md) · **Branch:** `s12-24-audit-index-and-pointer-resets` · **PR:** [#610](https://github.com/mendahu/provenencia/pull/610)
**State:** ready

## Round 0 — implement
Head `b49b20d0`. Gate PASS. Report: parity subtests seeded a second cover / preferred pointer via raw SQL; superseded by F-3.

## Round 1 — review
Reviewers: CON, MNT, SEC+PRF. Skipped: UX (no `macos/App/**`, strings, or error codes).

| ID | Lens | Severity | Title | Decision | Evidence / reason |
| --- | --- | --- | --- | --- | --- |
| MNT-1 | MNT | blocker | `optionalPointer` bypasses `col` / `colJSON` | Fix now → F-1 | verified |
| CON-1 | CON | major | Plan missing `**Landed:**` note | Fix now → F-4 | verified |
| CON-2 | CON | minor | Honesty does not require every `BucketOptional` be `SET NULL` | Fix now → F-2 | verified |
| MNT-2 | MNT | minor | Optional honesty branch duplicates Remaining/Release checks | Fix now → F-2 | verified |
| MNT-3 | MNT | minor | Parity second row built with raw SQL the domain refuses | Fix now → F-3 | verified |
| MNT-4 | MNT | minor | Shared update-before-delete assert copied in two packages | Fix now → F-3 | same-pass |

Nits (logged): `MNT-5 · facets.go · re-read only side-effect cols; New nil for fk` · `MNT-6 · facets.go · empty else on NullString goes with F-1`

SEC+PRF: Clean (no findings).

### Fix plan 1
**F-1 — Encode optionalPointer columns through `col` / `colJSON`** (MNT-1; nits MNT-5, MNT-6)
**F-2 — Honesty: every BucketOptional is SET NULL; Remaining once** (CON-2, MNT-2)
**F-3 — Legal twin rows + shared assert helper** (MNT-3, MNT-4)
**F-4 — Landed note** (CON-1)

## Round 1 — fix
Head `d04be4c2`. Gate PASS. F-1–F-4 done. Deviation: `RequireUpdateBeforeDelete` matches field keys and `New` only (twins use different Artifact/likeness ids). Merged `origin/main` → `65963bdd`; gate PASS again.

## Round 2 — review
Reviewers: CON, MNT. Skipped: SEC+PRF (Clean round 1). Scope: `b49b20d0...d04be4c2`.

| ID | Lens | Severity | Title | Decision | Evidence / reason |
| --- | --- | --- | --- | --- | --- |
| MNT-1 | MNT | blocker | (re-check) | Resolved | typed `col` + `colJSON` |
| CON-1 | CON | major | (re-check) | Resolved | Landed note present |
| CON-2 | CON | minor | (re-check) | Resolved | honesty branch |
| — | CON/MNT | — | new blockers | none | both lenses Clean |

Nits (logged): `CON-3 · dbtest/audit.go · RequireUpdateBeforeDelete doc says deep-equals Fields but compares keys + New only`

## Result
Gate PASS · CI green · open: none · follow-ups: none · to user: — · CON miss: none
