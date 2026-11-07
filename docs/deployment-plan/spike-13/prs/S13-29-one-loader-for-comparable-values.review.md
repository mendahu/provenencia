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
| M-6 | major | The sweep missed `docs/ideas/`, `add-match-rule:15`, the `profiles.go` header, several names | **Fixed.** STEP-8 and the DoD sweep; the unused "canonlinks package doc" line was dropped |
| M-7 | minor | `MemoKey` exported only for tests; `LayerKind` dead | **Fixed.** D-4 (unexported key, `MemoSnapshotForTest` reads through `Load`), D-12 (`LayerKind` goes) |
| M-8 | minor | The shared reader returned `promote.ErrInvalid` and imported a write package | **Fixed.** D-3: `errNilGraph`; `loadInputs` keeps Promote's error |
| M-9 | minor | `use-catalog-session` does not cover memos | **Fixed.** Header: `.cursor/rules/catalog-queries.mdc` under Rules that bind; skill dropped |
| M-10 | nit | `graphalign.CanonEdge` is a third link shape | Logged: Out row and STEP-8 entry in `docs/ideas/structural-debt.md` |
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

## Round 2 (2026-11-07)

Maintainability (fresh): **Fits with changes** — every round-1 answer holds; one new blocker, M-11. Executability (fresh): **Not ready** — X-1, X-2 blockers, the rest cheap.

| ID | Severity | Finding | Answer |
| --- | --- | --- | --- |
| M-11 | blocker | `subjectcompare.Subject` and the `recordlinks` guard row count on S13-26 doing work its Planned plan does not do (flat `Member`, no `readerPackages` row) | **To user.** Open question 1: which PR lands first; recommendation S13-29 first, S13-26's D-2 embeds `subjectcompare.Subject` and its STEP-3 adds the row |
| X-2 | blocker | Who adds the `recordlinks` row was conditional on a package that does not exist | **Fixed in part.** D-11 has the `evidencelayer` row only; the `recordlinks` row's owner is M-11's question |
| X-1 | blocker | The DoD sweep hit historical plan docs and could never pass | **Fixed.** The sweep covers `.agents/`, `.cursor/rules/`, `docs/*.md`, `docs/ideas/`; plans and logs under `docs/deployment-plan/` are records and are not edited |
| X-3 | major | Moved bodies use `f.t`/`f.c`/`f.y`; `fixture`'s `t` was not named | **Fixed.** D-13: bodies read `r.T`, `r.C`, `r.y`; `fixture` keeps `t`, `c`, `source`, `artifact` |
| X-4 | major | `gofmt -r` cannot match variadic `cite` and is one pattern per arity | **Fixed.** D-13: one GNU sed over the five files, then `go vet` names any non-fixture hit |
| X-5 | major | T-1 named `vcand`, which builds a Candidate, not `Values` | **Fixed.** T-1 builds values with `subjectprofile.Value(v, neg, "")` |
| X-6 | major | `Load` on an unknown graph was unspecified | **Fixed.** D-3: `ErrNoGraph`, nothing stored; T-9; Edges "Removed inputs" |
| X-7 | major | About 20 `graphalign.Bridge` literals in the tests; the check missed bare `Bridge{` | **Fixed.** STEP-3 names them; the check `rg`s `\bBridge\{` too |
| X-8 | minor | The local `primary` map's content | **Fixed.** D-10: ID, Ref, Kind only |
| X-9 | minor | STEP-7 / STEP-8 cross-references | **Fixed.** DoD and this log |
| X-10 | minor | `memoJSON` skips the unexported `byID` | **Fixed.** D-12 says so (`byID` is derived from `Subjects`) |
| M-5 gap 1 | minor | Evidence-graph tests blank-import `graphfill` | **Fixed.** D-13: `evidencelayertest` carries it |
| M-5 gap 2 | minor | `fixture` keeps `c`, `source`, `artifact` beside `C`, `Source`, `Artifact` | **Declined.** Renaming several hundred reads in five files buys no reader anything; the duplicates are set once in `newFixture` |
| M-5 gap 3 | minor | `New` copies `graphcachetest.OpenCatalog` | **Declined.** `New` also seeds a user, the vocabulary, a Source and an Artifact; sharing only the two-line open is not worth a cross-package dependency |
| M-6 gap | major | `add-match-rule` says nothing of `readerPackages` | **Fixed.** STEP-8 adds the sentence |
| M-7 side effect | minor | `MemoSnapshotForTest` now builds on a miss; the `LayerOK` check is dead | **Fixed.** D-12: doc rewritten, `LayerOK` and its check removed; T-3 carries the proof |
| Q4 | minor | `graphalign.Subject` alias and `subjectprofile.Subject` beside `subjectcompare.Subject` | Logged. The alias goes when `graphalign.Handle` folds in (structural-debt entry, STEP-8) |

Two rounds done. Open: M-11, to the user. The plan is Planned once it is answered.
