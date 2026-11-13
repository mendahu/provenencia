# Ship log — S15-36 A metadata board can hold a draft

**Plan:** [`S15-36-metadata-board-holds-a-draft.md`](S15-36-metadata-board-holds-a-draft.md) · **Branch:** `s15-36-metadata-board-holds-a-draft` · **PR:** #810
**State:** ready

## Round 0 — implement
Head `deab302cd`. Gate PASS (orchestrator rerun). Report deviations: a private `isRowRefusal(_:)` predicate in `MetadataBoardModel` shared by `classifySetFailure` and `rowRefusalMessage` (D-13's one check as one function; in the Landed note); the T-8 and T-10 tests land in later commits than their steps because they need the add row and `showRefusal`. §10 manual check not run by the implementer.

## Round 1 — review
Reviewers: CON, MNT, UX, SEC+PRF. UX routed by `macos/App/**`; SEC+PRF by `macos/App/Features/**` models (`MetadataBoardModel`, `MetadataDraftHost`) and the store writes in `GoStore`/`CoreInvoke`. Skipped: none.

SEC+PRF: no findings (reply rows come through `mapMetadataEntry` with Go-built `openHref`; draft host never builds a link from typed text; every pass is linear in one owner's rows).

| ID | Lens | Severity | Title | Decision | Evidence / reason |
| --- | --- | --- | --- | --- | --- |
| CON-1 | CON | major | Suggestion dismissed, re-added, then cleared in one draft leaves the board, but comes back as an empty suggestion after Save | Rejected by user (2026-11-13: "drop it") | verified: `MetadataDraftHost.set` appends `suggested: false`; engine `listWorkspaces` (`sourcemetadata.go:903-914`) marks every type suggestion `Suggested: true` and shows it empty unless its layout row is dismissed. Proposed fix: take `suggested` from the opened row. A suggestion dismissed before the form opened is not in `opened`, and its stored dismissed layout row survives `Apply`, so it stays consistent |
| MNT-1 | MNT | major | Draft host mirrors the engine's board rules with no parity fixture against Go | To user → Fix now → F-1 (user: shared JSON scenarios) | verified: no test ties `MetadataDraftHost` to `sourcemetadata`; CON-1 is such a drift, already present |
| UX-1 | UX | major | `showRefusal` discards typed text in an open row edit (its own or another row's) | To user → Fix now → F-2 (user: keep typed text) | verified: `beginEdit` calls `cancelEdit()` (reseeds the other row) and sets `drafts[fieldID] = entry.valueText`; T-10 never types before the refusal |
| UX-2 | UX | minor | Add dialog's picked field blanks with no message when another window adds it | Later PR → S15-34 map scope entry | needs a new string; S15-34 replaces the dialog with the inline row that will show it |
| MNT-2 | MNT | minor | "Row with a value for this field" lookup copied four times | Fix now → F-3 (same-pass cleanup) | verified: `beginEdit`, `requestClear`, `showRefusal`, `askClear` |
| MNT-3 | MNT | minor | `confirmAddRow` drops `add`'s result | Rejected | D-11 fixes `confirmAddRow() async` with no result and D-12 reads `addRow == nil`; the dialog goes in S15-34 |
| MNT-4 | MNT | minor | `sourcemetadata.invalid` test error built in four places | Fix now → F-4 (same-pass cleanup) | verified: `MetadataDraftHostTests`, `MetadataBoardModelTests`, `SourcePageModelTests` ×2 |
| CON-2 | CON | minor | Nothing pins that the Source page host confirms a clear | Fix now → F-5 | verified: `confirmsClear` only read on `TestMetadataBoardHost` in tests |
| CON-3 | CON | minor | AC-1.8 and AC-2.1 tests check half their rule | Fix now → F-6 | verified: no `draft == nil` after the drag back; no field-empty-value-typed case |

Nits (logged): `MNT-5 · MetadataDraftHost.swift:98-127 · CatalogWriteEffects() → .empty (D-7 says .empty)` · `CON-4 · MetadataDraftHostTests revertedDraftIsUnchanged · reordered(by:) check in its own test` · `UX-3 · SourcePageMetadataView.swift:31 · comment that the Task hop is intentional`

**To user, answered 2026-11-13:** CON-1 drop it (Rejected); MNT-1 shared JSON scenarios (F-1); UX-1 keep typed text (F-2).

### Fix plan 1
Supersedes: plan §5 "Rules this plan bends: None" (MNT-1, user decision) and D-13's "`beginEdit` … seeds the buffer" inside `showRefusal` (UX-1, user decision). Record both in the plan's Landed note.

**F-1 — Draft host parity scenarios shared by Go and Swift** (from MNT-1)
Files: `core/database/sourcemetadata/testdata/draft_board_scenarios.json` *(new)*, `core/database/sourcemetadata/sourcemetadata_test.go`, `macos/ProvenenciaTests/MetadataDraftHostTests.swift` (or a new `MetadataDraftHostParityTests.swift`), `.cursor/rules/macos-client.mdc` (one clause naming the scenario file beside the fixture-table line).
Failing test first: `TestDraftBoardScenarios` (Go) — fails until the file and runner exist.
Do: one JSON file of scenarios. Each: the type's suggestions (field ids, in order), the owner's starting values and layout rows (incl. a stored dismissed row), one or more actions (`set f v`, `clear f`, `dismiss f`, `reorder [ids]`), and the expected board after (each row: field id, value or none, suggested) and the expected `draft` (values, layout with dismissed flags) or null. Cover at least: AC-1.1's board; clear a suggested row (stays empty) and an extra row (leaves); dismiss an empty suggestion; dismiss on a valued row and on an extra (no-op); a new value goes after the layout rows; reorder; a value set back to its opened text (draft null); a stored-dismissed suggestion that stays out; adding a field a stored layout had dismissed. Do **not** include dismiss → re-add → clear of the same suggestion in one draft (CON-1, rejected by the user; known divergence) — put a one-line `"_note"` in the file saying so.
Go side: on a temp catalog, seed the type's suggestions and the owner's values/layout, then **for the expected board** run each action through the live writes (`Set`, `Clear`, `DismissSuggestion`, `Reorder`) and compare `ListWorkspace`; **for the draft** start again from the seed, `Apply` the scenario's expected draft, and compare `ListWorkspace` with the same expected board. Reuse the test catalog/seed helpers already in `sourcemetadata_test.go`.
Swift side: load the same file (locate it from `#filePath` up to the repo root; if the hosted test sandbox refuses the read, add it to the test target as a resource instead — and if neither works, stop and report), build `MetadataDraftHost` over the seed's board rows, drive the actions through `MetadataBoardModel`, and compare `host.entries` (field id, hasValue/valueText, suggested) and `host.draft` with the file.
Check: both tests pass; break one expected row by hand and both fail (then restore); scripts/check-pr.sh PASS.

**F-2 — A refusal keeps every typed buffer** (from UX-1)
Files: `macos/App/Features/Sources/MetadataBoardModel.swift`, `macos/ProvenenciaTests/MetadataDraftHostTests.swift`.
Failing test first: `refusalLandsOnItsRow` gains two cases — (a) Roll open with "13" typed (saved "12"), refusal on Roll → `drafts["roll"] == "13"`, Roll still in edit, error on Roll; (b) URL open with typed text differing from its value, refusal on Roll → `drafts[url]` keeps the typed text and URL stays in edit; Roll's error is placed (`fieldErrorID == "roll"`).
Do: `showRefusal` no longer calls `beginEdit` blindly. When the refused row is already in edit, keep its buffer. When another row is in edit and its buffer differs from its saved value, leave that edit open with its buffer and only place the error on the refused row. Otherwise open the refused row as today (`beginEdit`). Confirm the board view shows a row's `fieldError` while that row is not in edit; if it does not, say so in the report (S15-34 draws it) rather than changing views. Update `showRefusal`'s doc comment.
Check: the test passes; scripts/check-pr.sh PASS.

**F-3 — One saved-row lookup** (from MNT-2; same-pass cleanup)
Files: `MetadataBoardModel.swift`, `MetadataBoardDialogs.swift`.
Do: `func savedEntry(fieldID: String) -> CatalogMetadataEntry?` on `MetadataBoardModel`; use it in `beginEdit`, `requestClear`, `showRefusal`, `askClear`, and F-2's new check.
Check: `rg -n 'field\.id == \w+ && \$0\.hasValue' macos/App/Features/Sources` prints only `savedEntry`'s body; tests pass.

**F-4 — One refusal test error** (from MNT-4; same-pass cleanup)
Files: `macos/ProvenenciaTests/CatalogFixtures.swift`, `MetadataDraftHostTests.swift`, `MetadataBoardModelTests.swift`, `SourcePageModelTests.swift`.
Do: `catalogMetadataRefusal(params: [String]) -> CoreInvokeError` beside `catalogMetadataEntry`; move every site onto it.
Check: `rg -n '"sourcemetadata.invalid"' macos/ProvenenciaTests` prints only the helper.

**F-5 — The Source page host confirms a clear** (from CON-2)
Files: `SourcePageModelTests.swift`.
Do: beside the clear-confirm test (~line 1025), assert `await model.metadata.requestClear(fieldID:) == true` and that no clear was written.

**F-6 — Full AC-1.8 and AC-2.1 tests** (from CON-3)
Files: `MetadataDraftHostTests.swift`.
Do: `#expect(host.draft == nil)` after the drag back; in `addRowConfirmsOnlyWhenFilled` a case with `fieldID = ""`, `value = "F-7"`: `canConfirmAddRow` false and after `confirmAddRow()` no row is added.

**Ride-along (nit MNT-5):** `CatalogWriteEffects()` → `.empty` in `MetadataDraftHost` (D-7's wording).

**Later PR (UX-2):** written into S15-34's map scope entry (orchestrator).

## Round 1 — fix
Head `e783d1284`. Gate PASS (orchestrator rerun). F-1 – F-6 and the MNT-5 ride-along done. F-1's scenarios surfaced two divergences between the draft host and the engine, marked `known_divergence` in the scenario file (Swift skips that one comparison; Go always checks the engine): `suggested` (a field a stored layout had dismissed, added in the form, is an extra in the host and a suggestion in the engine — the CON-1 class) and `order` (the engine places a newly valued suggestion by fill time; the host keeps row order). F-2: `PVInlineEdit` draws a row error only while that row is in edit, so a refusal on a row not in edit is not drawn until S15-34 (Landed note).

Minors read by the orchestrator: F-3 (MNT-2) **Resolved** — `rg` prints only `savedEntry`'s body. F-4 (MNT-4) **Resolved** — only the `catalogMetadataRefusal` helper builds the code. F-5 (CON-2) **Resolved** — `sourcePageBoardAsksBeforeClearing`. F-6 (CON-3) **Resolved**. MNT-5 **Resolved** — no `CatalogWriteEffects()` left in the host.

## Round 2 — review
Reviewers: MNT (MNT-1 Fix now major), UX (UX-1 Fix now major). Not sent: CON (CON-1 rejected by the user; its minors read above), SEC+PRF (no findings). Scope `deab302cd...e783d1284`.
MNT round 2: MNT-1 **Resolved** (one scenario file run by Go `TestDraftBoardScenarios` through the live writes and `Apply`, and by Swift `MetadataDraftHostParityTests` through the host; rule documented in `macos-client.mdc`). No new findings.
UX round 2: UX-1 **Resolved** (both typed-buffer cases in `refusalLandsOnItsRow`).

Orchestrator check of the `order` divergence: Go's Apply check sorts both sides only for that scenario, and the draft sends the whole order, so after Save the engine shows the draft host's order. The only difference is between a live board and a draft board fed the same actions; neither board reorders after its own write. No user decision needed. The `suggested` divergence is the CON-1 class the user dropped.

| ID | Lens | Severity | Title | Decision | Evidence / reason |
| --- | --- | --- | --- | --- | --- |
| UX-4 | UX | blocker | A refusal placed on a row not in edit (another row holds typed text) is never drawn, yet `showRefusal` answers true, so the owner's Save fails silently | Fix now → F-7 | verified: `showRefusal` skips `beginEdit` when `hasTypedEdit` and returns true; `PVInlineEdit.restingBody` draws no `errorCaption`; S15-34's Next adds shows its Callout only on false |

### Fix plan 2
**F-7 — A refusal that cannot open its row goes to the owner's Callout** (from UX-4)
Files: `macos/App/Features/Sources/MetadataBoardModel.swift`, `macos/ProvenenciaTests/MetadataDraftHostTests.swift`, the plan's Landed note.
Failing test first: `refusalLandsOnItsRow`'s `typedElsewhere` case — URL open with typed text, refusal on Roll → `showRefusal` answers **false**, `editingFieldID == "url"`, `drafts["url"]` keeps the typed text, and `fieldError`/`fieldErrorID` are unchanged (nil, or what they were).
Do: in `showRefusal`, when another row is in edit with typed text (`hasTypedEdit`, `editingFieldID != fieldID`), answer false and change nothing, so the owner shows its general failure (Next adds: "shows its top Callout when false"), as AC-3.2's no-row cases do. The `typedHere` case (the refused row already in edit) and the plain case are unchanged: they answer true with the row in edit, where `PVInlineEdit` draws the error. Update `showRefusal`'s doc comment. Reword the Landed note's UX-1 line: drop "an error placed on a row that is not in edit is not drawn until S15-34"; say a refusal while another row holds typed text answers false and goes to the owner's Callout (review UX-4). No view or kit change.
Check: the test passes; `rg -n "placeRowError" macos/App/Features/Sources/MetadataBoardModel.swift` shows every call made with that row in edit; scripts/check-pr.sh PASS.

## Round 2 — fix
Head `c59a0c738`. Gate PASS (implementer). F-7 done: `showRefusal` answers false and changes nothing when another row holds typed text; orchestrator read the diff, and every `placeRowError` call is made with that row in edit. UX-4 **Resolved**. Merged `origin/main` (`96d5b1ff6`, #811) as `a2a11a5a9`; gate PASS (orchestrator rerun).

## Result
Gate PASS · CI pending · open: none · follow-ups: none · later PR: UX-2 → S15-34 map scope · to user: CON-1 (dropped), MNT-1 (shared JSON scenarios), UX-1 (keep typed text) — all answered 2026-11-13
