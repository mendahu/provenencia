# Ship log — S11-18 Contested marker on Promote rows

**Plan:** [`S11-18-contested-marker.md`](S11-18-contested-marker.md) · **Branch:** `s11-18-contested-marker` · **PR:** #507
**State:** review 2

## Round 0 — implement
Head `e25fb56a`. Gate PASS (orchestrator ran `scripts/check-pr.sh` against `d214e663`; 1236 Swift tests). Report deviations: the code is one commit, not one per behavior; the first docs commit `c76d9e0b` used a Status form plan-docs rejects and was pushed before its gate, fixed in `e25fb56a`; the PR body is shorter than §13 (rewritten at finish); no Landed note (no difference from the plan). Manual check (§10) not run.

## Round 1 — review
Reviewers: CON (Sonnet), MNT, UX, SEC+PRF. Skipped: none (`macos/App/**` and `Localizable.xcstrings` route UX; `macos/App/Features/Promote/PromoteModel.swift` and `PromoteFlow.swift` route PRF).

SEC+PRF: no findings. CON: one minor. MNT: two majors. UX: one minor.

| ID | Lens | Severity | Title | Decision | Evidence / reason |
| --- | --- | --- | --- | --- | --- |
| MNT-1 | MNT | major | Neighbor lookup copied into each move arm, the shape D-6 rejected | Fix now → F-1 | verified: `PromoteModel.swift:203,221,226,231` are four `flow.rows.first(where: { $0.subjectID == … })` lookups; same-pass cleanup also covers `PromoteFlow.swift:300,474`, `PromoteView.swift:434,845` |
| UX-1 | UX | minor | `.neighborPlaced` / `.neighborMoved` can name a Source ref as the record the neighbor matched | Fix now → F-1 | verified: arms read `neighbor.shownRef`, which falls back to the Source ref when the neighbor's target is not a handle (in-flight retarget, or a neighbor released later in settling) |
| MNT-2 | MNT | major | `noteDuplicates` still pattern-matches the handle ref beside the new `Target.handleRef` | Fix now → F-2 | same-pass cleanup: verified `PromoteFlow.swift:467,476`; supersedes D-6's "not touched" for these two sites (no visible change). The title-binding sites and `targetColumn`'s exhaustive switch stay |
| CON-1 | CON | minor | No Landed note on the plan | Fix now → F-4 | DoD line; written after F-1 – F-3, which are differences from the plan |
| CON-N1 | CON | nit → minor | The three wording tests build the same subjects and store | Fix now → F-3 | same-pass cleanup: this PR adds two copies of the setup in `PromoteModelTests` |

Nits (logged): `MNT-3 · PromoteModel.swift:208 · single-use local ref` (rides along in F-1) · `UX-2 · PromoteView.swift:492 · narrow-window pass in manual check step 1` (added to the PR body's manual check) · outside-lens notes from SEC+PRF and MNT (sheet subtitle and neighbor-group ref) are S11-20's, already on its scope entry.

### Fix plan 1

**F-1 — One row lookup, and move sentences name only a matched neighbor's record** (from MNT-1, UX-1, MNT-3)
Files: `macos/App/Features/Promote/PromoteFlow.swift`, `PromoteModel.swift`, `PromoteView.swift`, `macos/ProvenenciaTests/PromoteModelTests.swift`, `PromoteFlowTests.swift`.
Failing test first: in `PromoteModelTests.movesAreWordedNamingTheNeighbor`, add: with Ada's row retargeted to Skip (`model.flow.setTarget(subjectID: "a", token: "skip")` before the asserts, or a second load where Ada's proposal row is `target: "skip"`), `reasonText(.neighborPlaced(neighborID: "a"), on: bea)` and `.neighborMoved` both equal `L10n.string(L10n.Promote.reasonMoveUnknown)`, and `.neighborReleased` still equals `L10n.Promote.neighborReleased(neighbor: "Ada")`. Today they return the sentence with `CPR-A`.
Do:
- Add `func row(_ subjectID: String?) -> Row?` to `PromoteFlow`, doc "The row for this Subject, or nil when it is not on the page", beside `connectionLines`/`resolve` (non-mutating; `subjectID` nil → nil).
- Move every `rows.first(where: { $0.subjectID == … })` / `rows.first { $0.subjectID == … }` read onto it: `PromoteFlow.resolve` (line 300), `noteDuplicates` (line 474), `PromoteView.swift:434` (`model.flow.row(model.flow.sheetSubjectID)`), `PromoteView.swift:845`, and `PromoteModel.reasonText`. The `firstIndex` sites (mutations) and `rows.filter` at line 246 stay.
- In `reasonText`, join the three move cases into one arm `case .neighborPlaced(let id), .neighborMoved(let id), .neighborReleased(let id):` with one `guard let neighbor = flow.row(id) else { return L10n.string(L10n.Promote.reasonMoveUnknown) }`, then an inner `switch reason`: placed → `neighbor.target.handleRef.map { L10n.Promote.neighborPlaced(neighbor: neighbor.name, ref: $0) } ?? L10n.string(L10n.Promote.reasonMoveUnknown)`; moved → the same with `neighborMoved`; released → `neighborReleased(neighbor: neighbor.name)`; the inner switch ends `default: return L10n.string(L10n.Promote.reasonMoveUnknown)` with the comment `// Only the three move reasons reach this arm.` `.via` uses `flow.row(neighborID)` and passes `neighbor.shownRef` inline (MNT-3).
Check: `rg -n -F '$0.subjectID ==' macos/App/Features/Promote` prints only the `firstIndex` lines (PromoteFlow.swift setTarget, togglePin and two siblings) and the `rows.filter` at `mapped`; XT `PromoteModelTests` and `PromoteFlowTests` pass.

**F-2 — `noteDuplicates` reads `Target.handleRef`** (from MNT-2)
Files: `macos/App/Features/Promote/PromoteFlow.swift`.
Do: line 467 `guard let ref = rows[index].target.handleRef else { continue }`; line 476 `if let ref = other.target.handleRef {`. Leave `PromoteView.swift:640, :740` (they bind `title`) and `targetColumn`'s exhaustive switch.
Check: `rg -n "case \.handle\(_, let ref, _\)" macos/App/Features/Promote` prints only `PromoteFlow.swift:31` (the accessor) and `PromoteView.swift`'s `targetColumn`; XT `PromoteFlowTests` passes (its duplicate-note tests).

**F-3 — One setup for the wording tests** (from CON-N1)
Files: `macos/ProvenenciaTests/PromoteModelTests.swift`.
Do: add a private helper `loadedAdaAndBea(_ ada: CatalogPromoteGraphAlignmentRow, _ bea: CatalogPromoteGraphAlignmentRow) async -> PromoteModel` that builds subjects Ada and Bea, the store with the `Full name` property, `promoteProposals`, the model, and awaits `load()`. Move `reasonsAreWordedWithLabelsAndRefs`, `movesAreWordedNamingTheNeighbor` and `theContestedBandIsSpoken` onto it; each test reads its rows with `model.flow.row("a")!`.
Check: XT `PromoteModelTests` passes with the same assertions.

**F-4 — Landed note** (from CON-1)
Files: `docs/deployment-plan/spike-11/prs/S11-18-contested-marker.md`.
Do: under the Status line, add `**Landed:**` listing the differences from the plan: F-1 adds `PromoteFlow.row(_:)` and moves every row lookup onto it; the placed and moved sentences fall back to "after a connected row changed" when the neighbor has no record (UX-1), beyond D-5; F-2 moves `noteDuplicates` onto `Target.handleRef`, superseding D-6's "not touched" for those two sites; remove the "Logged, not done" sentence from §6 Next adds. Code is one commit, not one per behavior.
Check: `python3 scripts/check-plan-docs.py docs/deployment-plan/spike-11/prs/S11-18-contested-marker.md` reports no errors.

## Round 1 — fix
Head `a64615b9`. Gate PASS (orchestrator). F-1 – F-4 done. F-1's new assertions failed first against `65f29c07` ("after Ada → CPR-A was matched"), then passed. Deviations: the test retargets through `model.setTarget` (the flow's `rows` is `private(set)`); F-1 and F-3 share a commit; orchestrator moved the Landed note back inside the plan header (`a64615b9`).
Minors checked by the orchestrator: UX-1 **Resolved** (`PromoteModel.reasonText` placed/moved arms read `neighbor.target.handleRef` with the `reasonMoveUnknown` fallback); CON-1 **Resolved** (Landed note under Status); CON-N1 **Resolved** (`loadedAdaAndBea`).

## Round 2 — review
Reviewers: MNT (had Fix-now majors MNT-1, MNT-2). Scope `e25fb56a...a64615b9`. Not sent: CON, UX (minors only, checked above), SEC+PRF (no findings).
