# Ship log — S11-35 Hold a record Promote did not offer

**Plan:** [`S11-35-hold-found-record.md`](S11-35-hold-found-record.md) · **Branch:** `s11-35-hold-found-record` · **PR:** #543
**State:** ready

## Round 0 — implement
Head `811db8e`. Gate PASS (orchestrator re-run). PR #543 (draft). Report deviations: a private Swift test helper `unmatched(_:)` (recorded in the Landed note); `promotealign/evidence.go` moved on `main` since the Planned commit (#541) but is not a Current state row; `namedPersonOnNewSource` takes `personName` because `name` is already a local.

## Round 1 — review
Reviewers: CON (Sonnet), MNT, UX, SEC+PRF. Skipped: none. UX: `macos/App/Features/Promote/**`. SEC+PRF: `api/ffi/**` tests and a `macos/App/Features/**` model.

| ID | Lens | Severity | Title | Decision | Evidence / reason |
| --- | --- | --- | --- | --- | --- |
| MNT-1 | MNT | major | `hold` is the third hand-built menu `Alternative` | Fix now → F-1 | verified: `rg -n "Alternative\(" macos/App/Features/Promote` gives three constructions (`PromoteFlow.swift` `hold`, and two in `menu(for:kind:orders:)`), each computing sorted title, natural title and subtitle. Same-pass cleanup: this PR adds the third copy. Supersedes D-5's rejected "third `headerTitle` overload" (the builder replaces the four wrappers, it does not add one) |
| MNT-2 | MNT | major | `headerParts` is a second split of `CatalogConclusionHeader` | Fix now → F-2 | verified: `CatalogConclusionDetail.personHeader` / `eventHeader` / `placeHeader` (`GenealogyStore.swift:773-783`) do the same `if case` split, used by the three detail views. Codebase-fit row for `headerParts` missed them. Same-pass cleanup; supersedes D-5's `headerParts` |
| MNT-3 | MNT | minor | given/surname `NameFormatOrders` fixture copied again | Fix now → F-3 | verified: new block in `PromoteFlowTests.swift` (in `aFoundRecord…`) matches lines 111-120; more copies in other test files. Same-pass cleanup |
| UX-1 | UX | minor | A held record reads "No match" beside "Your choice" | Later PR → S11-D3 / S11-36 | the map already gives "how a held record with band None reads" to S11-D3 (`deployment-plan.md:176`); F-4 names the exact case there and in S11-36's scope entry |
| SEC-out-1 | SEC→CON | minor | A refused held pair keeps the row decided, and every re-proposal fails | Rejected | same path and behavior as `setTarget` on an offered handle that is merged later: `proposeAfterDecision` surfaces `proposeError` and the researcher changes the row. S11-36's picker searches only the row's kind (map S11-36 scope), so the wrong-kind case does not arise. No new code path |

Nits (logged): `CON-1 · macos/ProvenenciaTests/PromoteFlowTests.swift:894 · unmatched(_:) helper, recorded in the Landed note` · `MNT-4 · api/ffi/handlers/promote_align_test.go:31 · personOnNewSource doc could say it is the "Mary Smith" case of namedPersonOnNewSource` · `UX-2 · macos/App/Features/Promote/PromoteView.swift:429 · header: nil labels read "PER-7  PER-7"; for S11-36 to watch`

### Fix plan 1

**F-1 — One builder for every menu `Alternative`** (from MNT-1)
Files: `macos/App/Features/Promote/PromoteFlow.swift`.
Failing test first: none (refactor; behavior pinned by `aPersonRecordReadsNaturallyInTheSheetAndSortedInTheMenu`, `aFoundRecordIsHeldAndJoinsTheMenu` and the rest of `PromoteFlowTests`).
Do: add one `private static func item(id: String, ref: String, kind: EvidencePrimaryKind?, person: CatalogPersonHeader?, event: CatalogEventHeader?, place: CatalogPlaceHeader?, orders: NameFormatOrders, assessment: Assessment, reason: Reason) -> Alternative` that calls `conclusionTitle` (style `.sorted` → `title`, `.natural` → `readingTitle`, `fallback: ref`) and `conclusionSubtitle`. Make both appends in `menu(for:kind:orders:)` and the new-item branch of `hold` call it (the proposal row passes `EvidencePrimaryKind(rawValue: proposal.kind)` and `proposal.handleRef`, exactly as `headerTitle(_ row:)` does now). Delete the four `headerTitle` / `headerSubtitle` wrappers.
Check: `rg -n "Alternative\(" macos/App/Features/Promote` prints only the line inside `item(...)`; `rg -n "headerTitle|headerSubtitle" macos/App` prints nothing; scripts/check-pr.sh PASS.

**F-2 — One split of `CatalogConclusionHeader`** (from MNT-2)
Files: `macos/App/Platform/GenealogyStore.swift`, `macos/App/Features/Promote/PromoteFlow.swift`, `macos/App/Features/Conclusions/{Person,Event,Place}DetailContent.swift`, `macos/ProvenenciaTests/CatalogQueryRegistryTests.swift`.
Failing test first: none (refactor).
Do: add `extension CatalogConclusionHeader { var person: CatalogPersonHeader?; var event: CatalogEventHeader?; var place: CatalogPlaceHeader? }` (each `if case … else nil`) beside the enum. Delete `CatalogConclusionDetail.personHeader` / `eventHeader` / `placeHeader`; move their callers to `detail.header?.person` / `.event` / `.place` (the three detail views and `CatalogQueryRegistryTests.swift:530`). In `hold`, pass `header?.person`, `header?.event`, `header?.place` to F-1's builder; delete `headerParts`.
Check: `rg -n "headerParts|personHeader|eventHeader|placeHeader" macos` prints nothing; `rg -n "case \.(person|event|place)\(let" macos/App` prints only the extension's three lines plus switches that do more than extract the payload (list them in the report); scripts/check-pr.sh PASS.

**F-3 — One given/surname name-order fixture** (from MNT-3)
Files: `macos/ProvenenciaTests/PromoteFlowTests.swift`, plus every test file with an exact copy.
Failing test first: none (test refactor).
Do: add a test-target fixture `extension NameFormatOrders { static let givenSurname = NameFormatOrders(natural: [given "", surname ""], sorted: [surname ", ", given ""]) }` in an existing shared test-helper file if one exists, otherwise a new `macos/ProvenenciaTests/NameFormatOrdersFixtures.swift` (wired into the ProvenenciaTests target per add-swift-test). Replace every exact copy of that eight-line natural/sorted pair with `.givenSurname`. Leave variants (for example the surname-prefix one in `ObservationValueDisplayTests.swift`) as they are.
Check: `rg -n -A8 "NameFormatOrders\(" macos/ProvenenciaTests` shows no remaining exact given/surname pair outside the fixture; list any variants left and why; scripts/check-pr.sh PASS.

**F-4 — Name the band-None held case for S11-D3 and S11-36** (from UX-1; docs only)
Files: `docs/deployment-plan/spike-11/deployment-plan.md`.
Do: in the S11-D3 row and the S11-36 scope entry, say the case exactly: a held found record reads band None as "No match" beside "Your choice", and can also be Contested after Go answers (S11-35 T-2); S11-D3 draws what replaces "No match" for it and S11-36 ships it.
Check: `scripts/check-plan-docs.py docs/deployment-plan/spike-11/deployment-plan.md` 0 errors.

In the plan's Landed note, record that F-1 and F-2 supersede D-5 (no `headerParts`; the builder replaces the wrappers) and F-3 adds a shared test fixture.

## Round 1 — fix
Head `68bc4c6`. Gate PASS (orchestrator re-run). F-1 done (`rg "Alternative\("` in Promote prints only the builder; the wrappers are gone). F-2 done (`headerParts` and the `CatalogConclusionDetail.*Header` accessors are gone; the remaining `case .person(let` switches all map to titles or proto). F-3 done (`NameFormatOrders.givenSurname` in the new `NameFormatOrdersFixtures.swift`; the surname-prefix and surname-first variants are kept). F-4 done: UX-1 **Resolved** (orchestrator read the S11-D3 row and the S11-36 Ships line in `deployment-plan.md`).

## Round 2 — review
Reviewers: MNT (had the Fix-now majors MNT-1 and MNT-2). Not sent: CON, UX, SEC+PRF (no Fix-now blocker or major).
Scope `811db8e...68bc4c6`. MNT-1 **Resolved**, MNT-2 **Resolved** (reviewer). MNT-3 **Resolved** (orchestrator and reviewer: `NameFormatOrders.givenSurname`; no exact copy left). No new findings, no nits. No Fix plan 2.

## Result
Gate PASS · CI green (macOS first run; Go on re-run: the first run missed `TestLargeSourceBudgets/cold` by 15 ms, 2.515 s against 2.5 s, with no Go production change on this PR) · open: none · follow-ups: none · later PR: UX-1 → S11-D3 / S11-36 (written into the map) · to user: — · CON miss: none
