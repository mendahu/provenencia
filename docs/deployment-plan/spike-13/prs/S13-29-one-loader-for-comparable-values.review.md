# Review — S13-29 One loader for a Subject's comparable values

Plan: [`S13-29-one-loader-for-comparable-values.md`](S13-29-one-loader-for-comparable-values.md)

## Round 1 (2026-11-07)

Maintainability (fresh context, strongest model): **Fits with changes**. Executability (fresh context, implementer model): **Not ready** (X-1, X-2).

| ID | Severity | Finding | Answer |
| --- | --- | --- | --- |
| M-1 | blocker | T-2 added a second import walker beside `matchregistry`'s `checkPure` | **Fixed.** D-11: `readerPackages` table and `checkForbids` in `core/matchregistry/guard_test.go`, rows for `evidencelayer` and `recordlinks`; S13-27 adds `persondetail` |
| M-2 | major | `terms` was a caller input under a memo key that does not carry it | **Fixed.** D-3: `Load(g, graphID, namePattern)`; the build calls `canonlinks.LoadTerms` on a miss |
| M-3 | major | A third identical `{ID, Ref, Kind, Values}` struct and a per-call copy | **Fixed.** D-2: `subjectcompare.Subject`, `graphalign.Subject` an alias; D-8 passes the memo's slice. `graphalign.Handle` and `recordlinks.Member` are Out (structural-debt idea; S13-26's call) |
| M-4 | major | Metas are catalog-wide data in a per-graph memo; `LoadMetas`/`Decode` misplaced | **Fixed.** D-5: `valuerules.Metas`, `Layer` has no `Metas`. `Decode` stays in `evidencelayer`, with the package doc saying why the canon side reads it |
| M-5 | major | The loader's tests stay in Promote through a forwarding hook | **To user → answered 2026-11-07: move the fixtures in this PR, into a shared place.** D-13: `core/database/evidencelayer/evidencelayertest` (`evidencegraphstest` would cycle); D-12: the two loader tests move to `evidencelayer`, the hook goes; STEP-1, STEP-6 |
| M-6 | major | The sweep missed `docs/ideas/`, `add-match-rule:15`, the `profiles.go` header, several names | **Fixed.** STEP-7 and the DoD sweep; the unused "canonlinks package doc" line was dropped |
| M-7 | minor | `MemoKey` exported only for tests; `LayerKind` dead | **Fixed.** D-4 (unexported key, `MemoSnapshotForTest` reads through `Load`), D-12 (`LayerKind` goes) |
| M-8 | minor | The shared reader returned `promote.ErrInvalid` and imported a write package | **Fixed.** D-3: `errNilGraph`; `loadInputs` keeps Promote's error |
| M-9 | minor | `use-catalog-session` does not cover memos | **Fixed.** Header: `.cursor/rules/catalog-queries.mdc` under Rules that bind; skill dropped |
| M-10 | nit | `graphalign.CanonEdge` is a third link shape | Logged: Out row and STEP-7 entry in `docs/ideas/structural-debt.md` |
| X-1 | blocker | `matchregistry` imports `graphalign`, so the transitive guard could never pass | **Fixed.** Verified (`core/matchregistry/registry.go`, `alignment.go`). D-11 checks direct imports; AC-1.2 restated |
| X-2 | blocker | STEP-3's check could not pass before STEP-4 changed the test helper types | **Fixed.** The type changes (STEP-2) now come before the move, with the helper edits in the same step |
| X-3 | major | `LayerKind` fill was unspecified | **Fixed** by M-7: `LayerKind` is deleted |
| X-4 | major | Nil-graph order in `Load` | **Fixed.** D-3 puts the nil check first; today's order was safe too (`EvidenceGraphMemo` is nil-safe) |
| X-5 | major | The internal shape of `build` and `links` was left open | **Fixed.** D-10 |
| X-6 | major | `LayerBridges` now goes through the memo | **Fixed.** D-12 says so and why it is fine |
| X-7 | major | STEP-7 paths and a second `graphalign.Bridge` mention | **Fixed.** Full paths; both mentions named |
| X-8 | minor | The sweep was ambiguous for names that survive | **Fixed.** The DoD sweep lists qualified names and paths |
| X-9 | minor | `check-pr.sh` needs a Mac; how to get the `main` run; the guard's module root | **Fixed.** DoD says "run on a Mac"; STEP-6 names a worktree; STEP-5 names `checkPure`'s root lookup |
| X-10 | minor | T-1 fixture | **Fixed.** T-1 cites `helpers_test.go` `vcand` |

Round 2: due (round 1 had blockers and reshaped D-2, D-3, D-5; M-5's answer added D-13).
