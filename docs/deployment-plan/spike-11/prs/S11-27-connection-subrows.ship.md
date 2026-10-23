# Ship log — S11-27 Connection subrows on Promote

**Plan:** [`S11-27-connection-subrows.md`](S11-27-connection-subrows.md) · **Branch:** `s11-27-connection-subrows` · **PR:** #530
**State:** fixing 3 (user change)

## Round 0 — implement
Head `17034618`. Gate PASS (orchestrator re-ran). Draft PR #530. Report: no deviations from Decisions or Scope; three small differences recorded in the plan's Landed note (`isSymmetric` takes `EvidenceGraphPlacedBridge` after a rename on `main`; `PromoteConnectionSubrows` also takes the host ref for the expander identifier; two test helpers `link` and `mapped`). AC-3.2 – 3.4 and AC-4.4 are built but need the manual check.

## Round 1 — review
Reviewers: CON (Sonnet), MNT, UX, SEC+PRF. Skipped: none. UX: `macos/App/**` and `Localizable.xcstrings`. SEC+PRF: performance paths (`macos/App/Features/**` models — `PromoteFlow`, `PromoteModel` — and per-row list and graph hosting code).

| ID | Lens | Severity | Title | Decision | Evidence / reason |
| --- | --- | --- | --- | --- | --- |
| UX-1 | UX | major | Pressing a pointer scrolls but tells VoiceOver nothing | Fix now → F-1 | verified: `showHosted` (`PromoteView.swift:264-278`) only scrolls and sets `highlighted`; no announcement or focus move. D-12 fixes the scroll and visual highlight; an announcement adds to it and does not contradict it |
| UX-2, MNT-4 | UX, MNT | major | "+N more" and "Show fewer" are two buttons, so focus drops on toggle | Fix now → F-2 | verified: if/else builds two `PVButton`s (`PromoteView.swift:576-586`). D-11's strings are kept; only identity changes |
| MNT-1 | MNT | major | `Bridge` copies `BridgeFact` field by field; this PR widens both | Fix now → F-3 | verified: `kind` and `symmetric` added to both structs and to `lined`. Same-pass cleanup; supersedes D-3's "copied by `lined(_:_:)`" mechanism (the fields themselves stay) |
| MNT-2 | MNT | minor | `BridgeFact.kind` defaults to `.relationship` only for test literals | Fix now → F-3 | verified: `PromoteFlow.swift:287`; only production builder sets it |
| PRF-1, MNT outside-lens | PRF | minor | `connections` resolves ends with a linear `row(_:)` scan, twice per body | Fix now → F-4 | verified: `row(_:)` is `rows.first`; `connections` is walked by `connectionHosts` and `connectionCount` in the same pass. D-6 meant to close S11-26 PRF-1; the fix plan touches `PromoteFlow.swift` anyway |
| CON-1 | CON | minor | AC-1.9 "neither end is a row" has no test | Fix now → F-5 | verified: T-2 case (c) covers only anchors |
| UX-3 | UX | minor | A truncated subrow sentence can't be read with the mouse | Fix now → F-2 | verified: no `.help` on the sentence; neighbors use `.help(Text(verbatim:))` |
| UX-4 | UX | minor | Dropping `connectionsHint` leaves no cue that a connection is disputed on the Evidence graph | Rejected (user, 2026-10-23) | the connection was drawn by the researcher on this Source's graph and its sentence now sits under the row; a fix-it cue belongs on the subrow as the planned "open on Evidence graph" action. Was To user: the plan's Deletes list removes `connectionsHint` on purpose; restoring copy is a product call. Minor and independent of the fixes, so the loop continues and it goes to the user in the report |

Nits (logged): `CON-2 · PromoteFlowTests.swift · T-1 uses Set, no order asserted (T-5 covers it)` · `MNT-3 · PromoteView.swift subrow · 14 repeats PVMark size; make it a PromoteLayout constant` · `MNT-5 · PromoteView.swift subrow(_:) · use bridge.files` · `MNT-6 · PromoteFlowTests.swift · unset cases still hand-roll load beside mapped(...)` · `UX-5 · PromoteView.swift:510 · pointer needs .accessibilityInputLabels with visible title` · `UX-6 · PromoteView.swift:582 · "+1 more" reads with no noun`

### Fix plan 1

Order: majors, then the minors riding along. No blockers.

**F-1 — Announce where a pointer went** (from UX-1)
Files: `macos/App/Features/Promote/PromoteView.swift`. Skill: `add-localized-string` only if a new string is needed (it should not be).
Failing test first: none (view-only; no UI test harness). Manual check §10 step 6 covers it.
Do: in `showHosted(from:hosts:proxy:)`, after the scroll and highlight, post `AccessibilityNotification.Announcement(…).post()` with `L10n.Promote.connectionsHostedBy(name:)` filled with the target host row's display name (the same name the row shows). Do not move focus; D-12's scroll and visual highlight stay as they are.
Check: scripts/check-pr.sh PASS.

**F-2 — One expander button; readable truncated sentence** (from UX-2, MNT-4, UX-3)
Files: `macos/App/Features/Promote/PromoteView.swift` (`PromoteConnectionSubrows`).
Failing test first: none (view-only).
Do: replace the `if expanded { PVButton(fewerConnections) } else { PVButton(moreConnections(count:)) }` with one `PVButton` whose title is chosen by `expanded`, same `.ghost`, `.sm`, action and `accessibilityIdentifier`. Add `.help(Text(verbatim: bridge.sentence))` to the subrow sentence `Text`, as `PromoteEvidenceCards.swift:211` does.
Check: `rg -n "fewerConnections|moreConnections" macos/App/Features/Promote/PromoteView.swift` shows one `PVButton(` call for the expander; scripts/check-pr.sh PASS.

**F-3 — `Bridge` wraps `BridgeFact`; `kind` has no default** (from MNT-1, MNT-2)
Files: `macos/App/Features/Promote/PromoteFlow.swift`, `macos/App/Features/Promote/PromoteModel.swift`, `macos/ProvenenciaTests/PromoteFlowTests.swift`, and any other reader of `Bridge` fields (`rg -n "Bridge\b|\.bridge\." macos`).
Failing test first: none (refactor; existing T-1 – T-7 must stay green).
Do: `Bridge` becomes `@dynamicMemberLookup struct Bridge { var fact: BridgeFact; var state: ConnectionState; subscript<T>(dynamicMember key: KeyPath<BridgeFact, T>) -> T { fact[keyPath: key] } }` keeping its existing protocols (`Identifiable` via `fact.id`, `Equatable`, `Sendable`) and `files`. Move `name(of:)` onto `BridgeFact` if it only reads fact fields. `lined(_:_:)` becomes `Bridge(fact: bridge, state: state)`. Drop the `= .relationship` default on `BridgeFact.kind`; pass `kind` in the test literals (or build them with `link`). Supersedes D-3's "`Bridge` gains the same two fields, copied by `lined`": record that in the plan's Landed note. Call sites reading `bridge.sentence`, `bridge.endA` etc. keep compiling through the member lookup.
Check: `rg -n "symmetric:|kind:" macos/App/Features/Promote/PromoteFlow.swift` shows each field declared once on `BridgeFact` and no field-by-field copy in `lined`; tests pass; scripts/check-pr.sh PASS.

**F-4 — Resolve ends against a row index** (from PRF-1)
Files: `macos/App/Features/Promote/PromoteFlow.swift`.
Failing test first: none (performance; behavior unchanged, T-1 – T-6 stay green).
Do: in `connections`, build `[String: Row]` keyed by `subjectID` once (`Dictionary(rows.map { ($0.subjectID, $0) }, uniquingKeysWith: { first, _ in first })` so the first wins, as `rows.first` does today), and resolve both ends against it instead of calling `row(_:)` per end. Keep `row(_:)` for its other callers. Do not change `ConnectionHosts`' shape (D-6).
Check: `connections` no longer calls `row(_:)`; scripts/check-pr.sh PASS.

**F-5 — Test "neither end is a row"** (from CON-1)
Files: `macos/ProvenenciaTests/PromoteFlowTests.swift`.
Failing test first: `PromoteFlowTests/aConnectionFallsBackToTheEndThatIsARow` new case (d) — `link("x", .participation, "zz", "yy")` with neither id loaded: `hosted.isEmpty`, `pointers.isEmpty`, `connectionCount == 0`. (It may pass at once; that is fine, it is a coverage gap.)
Do: add the case.
Check: the test passes; scripts/check-pr.sh PASS.

## Round 1 — fix
Head `da02c54c`. Gate PASS (orchestrator re-ran). F-1 – F-5 done in one commit. Deviations: `Bridge.name(of:)` is a one-line forwarder (dynamic member lookup cannot forward methods); `resolve` takes `in rowByID:`; F-5 loads rows `xx`, `ww` so neither end is a row. Landed note records `Bridge` wrapping `BridgeFact`, superseding D-3's copy.
Minors read by the orchestrator: PRF-1 **Resolved** (`connections` builds `rowByID` once, no `row(_:)` call). CON-1 **Resolved** (new case in `aConnectionFallsBackToTheEndThatIsARow`). UX-3 **Resolved** (`.help` on the sentence). MNT-2 **Resolved** (`kind` has no default).

## Round 2 — review
Reviewers: UX (UX-1, UX-2), MNT (MNT-1). Not sent back: CON, SEC+PRF (minors only, resolved above).

| ID | Lens | Severity | Title | Decision | Evidence / reason |
| --- | --- | --- | --- | --- | --- |
| UX-1 | UX | major | (round 1) | Resolved | `AccessibilityAnnouncement.post(L10n.Promote.connectionsHostedBy(name:))` in `showHosted`; focus unchanged |
| UX-2, MNT-4 | UX, MNT | major | (round 1) | Resolved | one `PVButton`, title by `expanded`, same identifier |
| MNT-1 | MNT | major | (round 1) | Resolved | `Bridge` is `@dynamicMemberLookup { fact, state }`; fields declared once |

No new findings. Nits (logged): `MNT-7 · PromoteFlow.swift:332 · lined(_:_:) only forwards to Bridge(fact:state:); inline it`

## Result
Gate PASS · CI green · open: none · follow-ups: — · to user: UX-4 answered (dropped) · CON miss: none

## Round 3 — user change (2026-10-23)
After the report, the user asked to drop the "+N connections" pointer, its scroll and highlight, and the cap with "+N more" / "Show fewer", and to show each connection under both of its ends, all on this PR. Recorded in the plan as **Amended** A-1 – A-5. UX-4 was answered (dropped) before this.

### Fix plan 3

**F-6 — Show each connection under every end that is a row** (A-1, A-2, A-3)
Files: `macos/App/Features/Promote/PromoteFlow.swift`, `PromoteModel.swift`, `macos/App/Features/Catalog/Connect/EvidenceBridgeEdgeSummary.swift`, `macos/App/Platform/ConclusionVocabulary.swift`, `macos/ProvenenciaTests/PromoteFlowTests.swift`, `PromoteModelTests.swift`, `EvidenceBridgeEdgeSummaryTests.swift`.
Failing test first: rewrite `connectionsAreHostedByRule` as `connectionsShowUnderEveryEndThatIsARow`: participation John → Birth, location Birth → Toronto, place relationship Toronto → Ontario, relationship Mary → John (parent) and Ann → Bill (spouse), all ends rows → each connection's id is in `connectionsByRow` under **both** ends. Rewrite `aConnectionFallsBackToTheEndThatIsARow` as `aConnectionWithOneRowEndShowsOnlyThere`: the anchor case shows under John only with `OtherEnd(name: "b", filed: true)`; Promote on John alone shows under John with `filed: false`; both anchors, already filed, and neither end loaded → not in `connectionsByRow`. Keep `hostedConnectionsAreOrderedByStateThenSentence` (rename to `rowConnectionsAreOrderedByStateThenSentence`), reading `connectionsByRow`.
Do: A-2's `RowConnection` and `connectionsByRow` replace `ConnectionHosts`, `HostedConnection`, `Pointer`, `connectionHosts` and `ruleHost`. For each connection, for each distinct end that is a working row, append it with `otherEnd` set only when the other end is not a working row (D-5's note, unchanged). A-3: delete `kind` and `symmetric` from `BridgeFact` and `bridgeFacts`, both `isSymmetric` functions, `onlyAProductSymmetricRelationshipIsSymmetric`, `bridgeFactsCarryKindAndSymmetry` and the test-literal `kind:` / `symmetric:` arguments and the `link` helper's `symmetric` parameter.
Check: tests pass; `rg -n "ruleHost|ConnectionHosts|HostedConnection|connectionHosts|Pointer\b|pointers|isSymmetric|symmetric" macos` prints nothing.

**F-7 — Remove the pointer, highlight and cap from the view** (A-4, A-5)
Files: `macos/App/Features/Promote/PromoteView.swift`, `macos/App/Platform/L10n.swift`, `macos/App/Resources/Localizable.xcstrings`, `macos/ProvenenciaTests/PromoteEvidenceDisplayTests.swift`. Skill: `add-localized-string`.
Failing test first: in `connectionCopyIsWorded`, replace the `connectionsElsewhere` expectations with `L10n.Promote.connectionsOf(name: "John Smith") == "Connections of John Smith"`.
Do: `PromoteAlignmentRow` and `PromoteConnectionSubrows` take `[PromoteFlow.RowConnection]` and show all of them. Delete everything A-4 lists, including `ScrollViewReader`, `.id(row.subjectID)` and `reduceMotion` if nothing else in the file reads them. Subrow identifier `promote.row.<ref>.connection.<bridgeID>`; container label `connectionsOf(name:)`. Delete the four plural keys and `promote.connection.hostedBy`; add `promote.connection.of`.
Check: `rg -n "expandedHosts|highlighted|highlightGeneration|showHosted|subrowCap|connectionsElsewhere|moreConnections|fewerConnections|connectionsHostedBy|AccessibilityAnnouncement|onPointer|pointerCount" macos/App/Features/Promote macos/App/Platform/L10n.swift` prints nothing; `python3 scripts/check-localizable-xcstrings.py` OK.

**F-8 — Docs**
Files: `docs/deployment-plan/spike-11/prs/S11-27-connection-subrows.md` (Landed note only: add "Amended A-1 – A-5 shipped; supersedes the F-1 announcement, the F-2 expander and the earlier `Bridge`-wraps note on `kind`"), `docs/deployment-plan/spike-11/completed.md` (S11-27 paragraph: connections show under both ends that are rows; no pointer, cap or highlight), `docs/promote-graph-alignment.md` §3 "Bridges aren't rows" bullet (same), `docs/ideas/term-management.md` item 2 (remove the S11-27 sentence).
Do not edit the plan's Amended block. Check: `python3 scripts/check-plan-docs.py` 0 errors; scripts/check-pr.sh PASS.
