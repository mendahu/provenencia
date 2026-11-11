# Ship log — S15-06 Type suggestions have a level

**Plan:** [`S15-06-type-suggestions-have-a-level.md`](S15-06-type-suggestions-have-a-level.md) · **Branch:** `s15-06-type-suggestions-have-a-level` · **PR:** #782
**State:** ready

## Round 0 — implement
Base `main` @ `0cd37005`. Head `4580554c`. Gate PASS (orchestrator run: all nine checks). Report deviations: version bump 0.8.1 → 0.9.0 (main was 0.8.1, header allows); T-7 split across `TestAssignAndRemoveTypeField` and `TestRemoveTypeFieldDetachesOneLevel` (runRPC drives one handler per table); map "Ships" line for S15-06 still says the effects registry and catalog model read the level (D-7 left both unchanged) — not edited, noted under Landed; manual check §10 not run.

## Round 1 — review
Reviewers: CON, MNT, UX (`macos/App/**`), SEC+PRF (`api/ffi/**`, migration `000064.sql`, `core/database/**`). Skipped: none.
Conformance round 1 was dispatched twice: the first reviewer stopped on a tool-output size error before reading anything and returned no verdict; a fresh reviewer ran the lens in full.

| ID | Lens | Severity | Title | Decision | Evidence / reason |
| --- | --- | --- | --- | --- | --- |
| MNT-1, CON-4 | MNT, CON | major | Map "Ships" line and citation map § 6.6 "Readers of the join" still say the effects registry and catalog model read the level | Fix now → F-1 | verified: `deployment-plan.md:283`, `evidence-explained-citation-map.md:830` name both; D-7 leaves `effects/registry.go:460` and `catalogmodel/model.go` unchanged. Later PRs read these docs as binding |
| CON-1 | CON | major | AC-2.2 (one field listed at two levels at once) never asserted | Fix now → F-3 | verified: `TestRemoveTypeFieldDetachesOneLevel` returns from `reqFn` right after the two assigns and never lists `graph`; an assign that ignored the second level would pass |
| MNT-2 | MNT | minor | Proto and Swift comments for `suggested_field_count` / `TypeSuggestion` describe every level | Fix now → F-2 | verified: `engine.proto:340-347`, `GenealogyStore.swift:1184` unchanged; D-9 narrows the count to Source |
| CON-2 | CON | minor | AC-2.5 "nothing written" not tested for a refused remove | Fix now → F-4 | verified: `TestTypeSuggestionRefusesAnyOtherLevel` refuses calls on a type with no rows |
| CON-3 | CON | minor | Swift AC-2.8 test cannot tell the engine count from the reply length | Fix now → F-5 | the fake's count is the `.source` list length, equal to the reply |
| CON-6, MNT-4 | CON, MNT | minor | `Install` checks the type id inside the `FieldKeys` loop | Fix now → F-6 | verified `sourcevocab.go:186-191`: an entry with an unknown type and no fields passes silently; code this PR changed |

Security and performance: Clean, no findings. UX: Clean, no findings, no nits.

Nits (logged): `MNT-3 · core/database/sourcetypes/sourcetypes.go:52 · sqlSuggestedFields fragment's bind order is a hidden contract` · `MNT-5 · api/ffi/handlers/source_defs_test.go · raw level literals instead of string(vocabkey.Level…)` · `MNT-6 · api/ffi/handlers/source_defs_test.go:780-800 · listSuggestionLabels and suggestionIDs could share one helper` · `CON-5 · core/database/sourcetypes/sourcetypes.go:319 · SuggestedFields dropped the len(typeID) != 16 check (callers pass parsed ids)`

### Fix plan 1

**F-1 — Make the spike docs match D-7** (from MNT-1, CON-4)
Files: `docs/deployment-plan/spike-15/deployment-plan.md` (S15-06 "Ships" paragraph, line 283), `docs/deployment-plan/spike-15/evidence-explained-citation-map.md` (§ 6.6 "Readers of the join", line 830), `docs/deployment-plan/spike-15/prs/S15-06-type-suggestions-have-a-level.md` (Landed note).
Do: line 283 → "`sourcevocab` and `sourcetypes` read it (the effects registry and catalog model are unchanged, D-7); the suggestion RPCs take a level and return the type's Source-level field count, and the Source page's board reads `source` only." Line 830 → "filter by level: `sourcetypes` and `sourcevocab` (suggestion ops and Install). The `effects` registry and the catalog model are unchanged (D-7)." Drop the Landed bullet that records the stale line.
Check: `rg -n "catalog model" docs/deployment-plan/spike-15/deployment-plan.md docs/deployment-plan/spike-15/evidence-explained-citation-map.md` prints only lines that say "unchanged"; scripts/check-pr.sh PASS.

**F-2 — Say "at the Source level" on the wire and Swift count** (from MNT-2)
Files: `api/proto/engine.proto` (`SourceType.suggested_field_count`, `TypeSuggestion` comment), `macos/App/Platform/GenealogyStore.swift` (`CatalogSourceType.suggestedFieldCount` doc); regenerate `engine.pb.go` and `engine.pb.swift` with the repo's proto generation.
Do: `suggested_field_count` → "How many metadata fields this type suggests at the Source level — …"; `TypeSuggestion` → "a field this type suggests at one level, in its stored order"; same wording on the Swift doc.
Check: generated files regenerated, no other diff in them; scripts/check-pr.sh PASS.

**F-3 — Assert one field is listed at both levels before the remove** (from CON-1)
Files: `api/ffi/handlers/source_defs_test.go`, `core/database/sourcevocab/sourcevocab_test.go`.
Failing test first: not applicable (test gap; code is correct) — add the assertion and confirm it passes; then confirm it would fail by temporarily making the graph assign a no-op locally (do not commit that).
Do: in `TestRemoveTypeFieldDetachesOneLevel`'s `reqFn`, after the assign loop, list `graph` and `source` with `listSuggestionLabels` and require `[Folio]` at each. In `TestAppendSuggestion`'s per-level delete case, `ListSuggestions(c, typeID, vocabkey.LevelGraph)` before the delete and require the field.
Check: both tests pass; scripts/check-pr.sh PASS.

**F-4 — Prove a refused remove deletes nothing** (from CON-2)
Files: `api/ffi/handlers/source_defs_test.go` `TestTypeSuggestionRefusesAnyOtherLevel`.
Do: before the refused calls, assign `fieldID` at `"source"` with a valid request. At the end require `source` → `[Folio]` and `graph`, `artifact` empty.
Check: test passes; scripts/check-pr.sh PASS.

**F-5 — Let the Swift test tell the engine's count from the list length** (from CON-3)
Files: `macos/App/Platform/FakeStore.swift`, `macos/ProvenenciaTests/SourceTypesModelTests.swift`. Skill: `add-swift-test`.
Failing test first: in `assigningWritesAtSourceLevelOnly` (or a new sibling test), set a fake count override for the type that differs from the list length (e.g. 7) and expect the model's Fields column to show 7; it must fail if the model used `updated.suggestions.count`.
Do: add `var suggestedFieldCountOverride: [String: Int] = [:]` on `FakeStore`, read by the edit replies of `assignTypeField` and `removeTypeField` (falling back to today's `.source` count).
Check: test passes; scripts/check-pr.sh PASS.

**F-6 — Check an Install entry's type before its field loop** (from CON-6, MNT-4)
Files: `core/database/sourcevocab/sourcevocab.go` (`Install`, lines 186-195), `core/database/sourcevocab/sourcevocab_test.go`.
Failing test first: if `Install`'s seed list can be swapped in tests, a case with an unknown `TypeKey` and empty `FieldKeys` expects `ErrInvalid`; if the seed list is not injectable, skip the test and say so in the report.
Do: move `if len(typeID) == 0 { return ErrInvalid }` above the `FieldKeys` loop; keep the per-field `len(fieldID) == 0` check inside.
Check: `go test ./core/database/sourcevocab/`; scripts/check-pr.sh PASS.

## Round 1 — fix
Head `5945913e`. Gate PASS (orchestrator run: all nine checks). F-1 `77815a82`, F-2 `0e363bc1`, F-3 + F-4 `6654d5af` (the `sourcevocab_test.go` half of F-3 rides in `14fe27a1`), F-5 `5945913e`, F-6 `14fe27a1`. F-3: the new assertion fails with "graph list [] before the remove, want [Folio]" when the handler ignores the graph level (reverted, not committed). F-5: no failing-first possible (the model already reads the engine's count); the new test `fieldsColumnShowsTheEnginesCountNotTheReplyLength` would fail on `updated.suggestions.count`. F-6: the new `TestInstall` case failed before the fix with `Install = <nil>, want ErrInvalid`.

## Round 2 — review
Reviewers: CON (CON-1, CON-4 majors), MNT (MNT-1 major). Scope `4580554c...5945913e`. Not sent back: UX, SEC+PRF (clean in round 1).

| ID | Lens | Severity | Title | Decision | Evidence / reason |
| --- | --- | --- | --- | --- | --- |
| CON-1 | CON | major | AC-2.2 never asserted | Resolved | `TestRemoveTypeFieldDetachesOneLevel` lists `graph` and `source` before the remove; `TestAppendSuggestion` lists `graph` before the delete |
| MNT-1, CON-4 | MNT, CON | major | Spike docs contradict D-7 | Resolved | `deployment-plan.md:283`, `evidence-explained-citation-map.md:830` now say unchanged (D-7) |
| MNT-2 | MNT | minor | Count comments describe every level | Resolved (orchestrator read) | proto, generated files, `GenealogyStore.swift` say "at the Source level" / "at one level" |
| CON-2 | CON | minor | Refused remove not proven to delete nothing | Resolved (orchestrator read) | refusal test assigns at `source` first, then requires `[Folio]` there |
| CON-3 | CON | minor | Swift count test could not tell count from length | Resolved (orchestrator read) | `FakeStore.suggestedFieldCountOverride`; new test expects 7 with one row |
| CON-6, MNT-4 | CON, MNT | minor | Install type check inside field loop | Resolved (orchestrator read) | `sourcevocab.go:188` checks the type before the loop |

No new findings. Both verdicts Clean. No Fix plan 2.

Nits (logged): `MNT-7 · macos/App/Platform/GenealogyStore.swift:1184-1187 · rewrapped doc comment leaves "rows. Only" on a short line` · `MNT-8 · deployment-plan.md:283, evidence-explained-citation-map.md:830 · cite "S15-06 D-7" rather than bare "D-7" in spike-wide docs`

## Result
Gate PASS at `5945913e` · `main` unchanged since the branch was cut (`0cd37005`), no merge needed · CI green (localizable, test ×2) · open: none · follow-ups: none · to user: — · manual check §10 not run (user)
