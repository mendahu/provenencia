# Ship log — S11-40 Implied links in claim evidence

**Plan:** [`S11-40-implied-links-evidence.md`](S11-40-implied-links-evidence.md) · **Branch:** `s11-40-implied-links-evidence` · **PR:** #579
**State:** ready

## Round 0 — implement
Base `main` @ `b11d1cb`.
Head `e1642c9e`. Gate PASS (orchestrator run). PR #579 (draft).
Report deviations:
- STEP-4 and STEP-5 landed as one commit: reading `Relationships` makes the tree-marriage case show a line before pivots, so the old "AC-3.3" subtest had to flip in the same commit (STEP-4's "no test changed" check could not hold).
- `kinship.Shape` holds a `[]byte`, so tests compare shapes by `Pattern` or `reflect.DeepEqual`.
- Swift reads the pivot header through the generated oneof (`case .person(let h)?`); the generated type has no `hasPerson`.
- `marriageScene` test fixture carries a few fields beyond D-16's list (`treeAnn`, `treeRelationship`, `recordSource`, `recordArt`).
- AC-3.5 (manual check) not run: needs the app.

## Round 1 — review
Reviewers: CON (Sonnet), MNT, UX, SEC+PRF. Skipped: none. UX: `macos/App/Platform/**` and `engine.proto` changed. SEC: `api/ffi/**` and the matching layers (`core/kinship`, `core/claimevidence`). PRF: `core/database/**` (evidence query, pins, conclusion headers).

| ID | Lens | Severity | Title | Decision | Evidence / reason |
| --- | --- | --- | --- | --- | --- |
| PRF-1 | PRF | blocker | `evidence` reads `Relationships` for every handle row, places included; cost grows with the place's events | Fix now → F-1 | verified: `evidence.go:157` calls `q.kin.Relationships(row.HandleID)` with no kind check; `Relationships` → `loadAround` + `For` (`canonlinks/reader.go:340`), and `forSlot` drops on every handle commit |
| MNT-1 | MNT | major | Swift pivot header is two optionals beside `CatalogConclusionHeader` | To user → answered: use the enum → F-2 | contradicts D-13; user chose `pivot: CatalogConclusionHeader?` (same Person/Event header the Go side already reads through `conclusionheaders.ByIDs`); supersedes D-13's `pivotPerson`/`pivotEvent` |
| MNT-2 | MNT | major | `canonLinks` hand-writes "read a link from one end" beside `graphalign.Bridge.From` | Fix now → F-3 (same-pass cleanup) | verified: `evidence.go:437-446` switch over `l.A/l.B/KeyA/KeyB`; `graphalign.Bridge` is `A,B,KeyA,KeyB,Associations` + `kinship.Recording`, field for field `kinship.Link` |
| MNT-3 | MNT | major | `bridgeIndex` is a second index of the row graph's Observations by Subject beside `observationsOf` | Fix now → F-4 (same-pass cleanup) | verified: `evidence.go:100-103` and `:595-597` scan the same `eg.Observations` by `SubjectID` |
| CON-1 | CON | minor | AC-1.4 and AC-1.6 subtests weaker than their ACs | Fix now → F-5 | file this PR adds tests to |
| MNT-4 | MNT | minor | `kinship-shapes.md:67` still says "S11-40 decides" after stating the decision | Fix now → F-6 | `rg "S11-40 decides"` matches line 67 |
| UX-1 | UX | minor | An implied tree link's Card does not say how the tree recorded it | Later PR (S11-42) | already S11-42's scope: map scope entry "Implied links on the evidence sheet" words `canon.pattern` / pivot; plan Objective "the sheet's words do not change until S11-42" |

Nits (logged): `MNT-5 · core/claimevidence/build.go:77-80,117-120 · "clone first recording, merge the rest" fold written twice` · `MNT-6 · core/database/promotealign/evidence.go:116,492 · pivot kind read from in.primary on one side, ix.types on the other; say which is the source of truth` · `UX-2 · macos/App/Platform/GenealogyStore.swift:407 · "the event or parent between it passes through" → "the event or the parent in between"` · `CON-2 · Landed note · canon Pivot.Kind comes from the bridge index type map`

### Fix plan 1

**F-1 — Read kinship relationships only for person rows** (from PRF-1)
Files: `core/database/promotealign/evidence.go`, `core/database/promotealign/perf_test.go`, `core/database/promotealign/evidence_test.go`.
Failing test first: a `TestLargeSourceBudgets` subtest that runs evidence on the `onePlace` Source's place row after a handle commit, beside `births at one place warm`, asserting the evidence budget and that the place row's `Relationships` is not read (count `kinshipSlot`/`For` lookups, or assert the query count stays at the person-row level). If the budget alone does not go red on `largeFixture`'s 6,000-event city, add a counting assertion that does.
Do: in `evidence`, call `q.kin.Relationships(row.HandleID)` only when the row's kind is `vocabkey.Person` (`in.primary[string(row.SubjectID)].kind`); otherwise `rels` stays nil. Kinship links have only person ends, so event and place rows lose nothing. `canonlinks` stays unchanged (plan Scope).
Check: the new subtest passes; every evidence test passes; scripts/check-pr.sh PASS.

**F-2 — One pivot header in Swift** (from MNT-1; supersedes D-13's two optionals, by the user's answer)
Files: `macos/App/Platform/GenealogyStore.swift`, `macos/App/Platform/GoStore.swift`, plan Landed note.
Do: `CatalogClaimEvidenceShape` drops `pivotPerson` / `pivotEvent` for `var pivot: CatalogConclusionHeader? = nil`. `mapEvidenceShape` switches the oneof into `.person(mapPersonHeader(h))` / `.event(mapEventHeader(h))`, as `mapSearchHeader` (`GoStore.swift`) does; nil oneof → nil. While there, reword the doc comment per UX-2 ("the event or the parent in between"). Record in the plan's Landed note that this supersedes D-13's `pivotPerson`/`pivotEvent`.
Check: `rg "pivotPerson|pivotEvent" macos` prints nothing; scripts/check-pr.sh PASS (xcode-test).

**F-3 — One "read a link from one end" on `kinship.Edge`** (from MNT-2; same-pass cleanup)
Files: `core/kinship/kinship.go`, `core/kinship/kinship_test.go`, `core/graphalign/types.go`, `core/graphalign/bridge_from_test.go`, every `graphalign.Bridge{` literal (`core/database/promotealign/{layer,evidence}.go`, `core/claimevidence/build.go`, `core/connectrules/connectrules.go`, and their tests, `core/database/subjectvocab/segment_section_test.go`, `core/graphalign/*_test.go`), `core/database/promotealign/evidence.go`, `core/database/canonlinks/reader.go`.
Failing test first: `TestEdgeFrom` in `core/kinship` (A end, B end, neither) — red until `Edge.From` exists.
Do: move `From` from `graphalign.Bridge` to `func (e Edge) From(id []byte) (other []byte, key vocabkey.BridgeKey, ok bool)` in `core/kinship/kinship.go`, same body. `graphalign.Bridge` embeds `kinship.Edge` in place of its own `A, B, KeyA, KeyB, Associations` (keep its doc comment on Associations on the struct) and drops its `From`; update the literals. Move `bridge_from_test.go`'s cases to `TestEdgeFrom` (delete the graphalign copy). In `canonLinks`, replace the `switch` with `other, key, ok := l.From(from.ID); if !ok || !bytes.Equal(other, to.ID) { continue }`. In `canonlinks/reader.go` (the `births` loop, `if bytes.Equal(imp.A, id) { add(imp.B) } else { add(imp.A) }`), use `other, _, _ := imp.From(id); add(other)`. This supersedes the plan's "`canonlinks` is read, not changed" for that one loop (same-pass rule).
Check: `grep -rnE "bytes\.Equal\([a-z]+\.(A|B), [a-z.]+\)" core --include=*.go | grep -v _test` prints only `canonlinks/reader.go` `atEnd` / the filter at :229 and `promotealign/canon.go:187` (membership tests, not end reads) and `kinship.go`'s `Edge.From`; scripts/check-pr.sh PASS.

**F-4 — One per-graph index in the evidence reader** (from MNT-3; same-pass cleanup)
Files: `core/database/promotealign/evidence.go`.
Do: rename `bridgeIndex` (type and method) to `graphIndex`, doc "a graph's Subjects by type and Observations by Subject, indexed once per request". Delete `observationsOf` in `evidence`; read the row graph's layer-link Observation ids from `q.graphIndex(in.graphID)` `.obs` (taking `o.ID`). Same order as today (graph order).
Check: `rg "observationsOf|bridgeIndex" core` prints nothing; every evidence test passes unchanged; scripts/check-pr.sh PASS.

**F-5 — Tighten AC-1.4 and AC-1.6 subtests** (from CON-1)
Files: `core/database/promotealign/evidence_test.go`.
Do: AC-1.6 asserts `r.Canon.Pattern == kinship.PatternParentOfParent` and `bytes.Equal(r.Canon.Pivot, beaH)`. AC-1.4 compares the bundle's last Observation to the birth's one `event_type` Observation id exactly.
Check: both subtests pass.

**F-6 — Drop the stale "S11-40 decides" sentence** (from MNT-4)
Files: `docs/deployment-plan/spike-11/kinship-shapes.md`.
Do: delete the trailing "Today a bundle is one bridge … so S11-40 decides the bundle's shape and its pin id." sentence.
Check: `rg "S11-40 decides" docs` prints nothing.

## Round 1 — fix
Head `1bcd2c73`. Gate PASS (orchestrator run). F-1 – F-6 done, one commit each.
- F-1: the query count is the same with and without the fix (nodes already cached), so the new subtest `TestLargeSourceBudgets/evidence on a place row` is time-based (`budgetPlaceEvidence` 10 ms, fastest of 3, cold graph): 28.9 ms before the fix, about 2 ms after.
- F-2: the plan's Landed note records that `pivot: CatalogConclusionHeader?` supersedes D-13's `pivotPerson`/`pivotEvent`; it also gained CON-2's sentence.
- F-3: `graphalign.Bridge` embeds `kinship.Edge`; `bridge_from_test.go` moved to `TestEdgeFrom`.
- The implementer's push failed (sandbox, no network); the orchestrator pushed.
- CON-1 (minor): **Resolved**. `8d8a75a1` compares AC-1.4's last Observation to the birth's one `event_type` id and asserts AC-1.6's `Canon` pattern and pivot.

## Round 2 — review
Reviewers: MNT (MNT-1 – MNT-3 majors), SEC+PRF (PRF-1 blocker). Not sent back: CON (minor only, resolved above), UX (no Fix-now item).

| ID | Lens | Severity | Title | Decision | Evidence / reason |
| --- | --- | --- | --- | --- | --- |
| MNT-1 | MNT | major | Swift pivot header | Resolved | `CatalogClaimEvidenceShape.pivot: CatalogConclusionHeader?`; `rg "pivotPerson\|pivotEvent" macos` empty |
| MNT-2 | MNT | major | "read from one end" copy | Resolved | `kinship.Edge.From` only; `graphalign.Bridge` embeds `kinship.Edge`; fix-plan grep prints only the expected lines |
| MNT-3 | MNT | major | second Observations index | Resolved | `graphIndex` only; `rg "observationsOf\|bridgeIndex" core` empty |
| PRF-1 | PRF | blocker | `Relationships` read for place rows | Resolved | `evidence.go:155-161` reads it only for person rows (`checkFixed` keeps the handle's kind equal to the Subject's); `evidence on a place row` 2.1 ms vs 28.9 ms before |

No new findings. No Fix plan 2.

## Result
Gate PASS · CI pending · open: none · follow-ups: none (UX-1 is S11-42's scope) · to user: MNT-1 (answered: use `CatalogConclusionHeader`, supersedes D-13) · CON miss: none
