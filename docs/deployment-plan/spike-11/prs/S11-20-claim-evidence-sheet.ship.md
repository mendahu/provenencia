# Ship log — S11-20 Claim evidence sheet

**Plan:** [`S11-20-claim-evidence-sheet.md`](S11-20-claim-evidence-sheet.md) · **Branch:** `s11-20-claim-evidence-sheet` · **PR:** #525
**State:** ready

## Round 0 — implement
Head `ec0414a2`. Gate PASS (orchestrator re-ran). PR #525 (draft). Report deviations: STEP-5–9 in one commit (sheet does not build between them); `L10n.Promote.pinRelationship` removed though D-12 said "kept" (replaced by `spokenRelationship`, no reader left); extra identifier `promote.sheet.section.<relationship id>.<property>.<n>`; "Pick a record" empty state above claim fields for New rows (Frame 13 order); neighbor ref and line badge hidden from VoiceOver (carried by checkbox label); `notCompared` reuses key `promote.sheet.unknown` with text "Not compared"; §10 manual check not run.

## Round 1 — review
Reviewers: CON (Sonnet), MNT, UX, SEC+PRF. Skipped: none. UX for `macos/App/**` and strings; SEC+PRF because `macos/App/Features/Promote/PromoteModel.swift` is a Features model (performance path); no security path touched, so the combined reviewer covers performance mainly.

| ID | Lens | Severity | Title | Decision | Evidence / reason |
| --- | --- | --- | --- | --- | --- |
| MNT-1 | MNT | major (reviewer: blocker) | Two lookups for the bridge between a row and its neighbor | Fix now → F-1 | verified: `PromoteModel.swift:252` inline `flow.bridges.first` end-set match against `PromoteEvidenceDisplay.bridge` (`:38-46`). Same-pass cleanup. Re-rated major: today both pick the same bridge, the risk is drift |
| UX-1 | UX | major | Member checkboxes under a non-pinnable line stay enabled; their pins are dropped on Done | Fix now → F-2 | verified: `EvidenceMemberRow` (`PromoteEvidenceSheet.swift:447-463`) builds `PVCheckbox` with no `isDisabled`; Go gives every member a pin id (`promote_evidence.go:128-130`); `ResolvePins` drops non-offered ids. Against D-7 |
| MNT-2 | MNT | major | `EvidenceLineRow` / `EvidenceLikenessLine` near-copies; card header, dispatch and outcome Badge pasted 2–4× | Fix now → F-3 | verified by reading both views; same-pass cleanup in code this PR adds |
| MNT-3 | MNT | major | Display choices left in view bodies, untested (D-2) | Fix now → F-4 | verified: `spoken`, `detail`, `linkText`, `memberSource`, `isEmpty`, `memberText` sit in views; tests only call L10n formatters. D-2 binds |
| UX-3 | UX | minor | Spoken labels break for an untitled Source; likeness falls back to "Likeness" | Fix now → F-4 | verified: `spokenMember(source: context.sourceTitle(pin) ?? "", …)` at `:424`, `:458` |
| MNT (outside) | UX | minor | `spoken` joins two localized sentences with `+ " " +` | Fix now → F-4 | verified `:424`; a translator cannot reorder |
| CON-1, UX-2 | CON, UX | minor | "Check every line…" intro shows while loading and on failure | Fix now → F-5 | verified: `else` branch at `:137` runs when `evidence == nil`; D-17 Loading lists no intro |
| CON-2 | CON | minor | Expanded likeness line lists members as the word "Likeness" | Fix now → F-6 | verified: `EvidenceMemberRow` has no `.depiction` branch; D-10 says a depiction draws the tile |
| MNT-4 | MNT | minor | `lineSentence` encodes value kind as three Bools; dead `.different` arm | Fix now → F-7 | code this PR adds |
| MNT-5, CON-3 | MNT, CON | minor | `EvidenceLayout.crop` repeats `LikenessTile.crop` | Fix now → F-8 | verified `:218` vs `LikenessTile.swift:353` |
| MNT-6 | MNT | minor | DesignSystem README Panel/Checkbox rows stale | Fix now → F-9 | verified `README.md:172,176` |
| MNT (outside) | CON/UX | minor | Sheet link falls back to "with <name>"; page row falls back to `bridge.sentence` | Rejected | D-14 binds the sheet's fallback to `relationshipWith(name)` |

Nits (logged): `CON-4 · PromoteModelTests theSheetMarksSpeakTheContestedBand · name overstates` · `MNT-7 · PromoteModel.swift:242 · neighbor(_:) only forwards to flow.row` · `MNT-8 · PromoteFlow.swift:227 · BridgeFact.kind default` · `MNT-9 · PromoteEvidenceSheet.swift:211 · EvidenceLayout.box repeats PVCheckbox 16 pt` · `UX-4 · :407-413 · untitled single member repeats the value` · `UX-5 · :523 · untitled likeness caption empty` · `UX-6 · :369-377 · only the box toggles a line`
SEC+PRF: clean, no findings.

### Fix plan 1

**F-1 — One bridge lookup** (from MNT-1)
Files: `macos/App/Features/Promote/PromoteEvidenceDisplay.swift`, `macos/App/Features/Promote/PromoteModel.swift`, `macos/ProvenenciaTests/PromoteEvidenceDisplayTests.swift`.
Failing test first: `PromoteEvidenceDisplayTests` `bridgeBetweenIgnoresKindWhenNoneIsGiven` — two bridges of different kinds between the same pair; `bridge(between:and:kind: nil, in:)` returns the first in page order, with a kind returns that kind's.
Do: add `static func bridge(between row: String, and neighbor: String?, kind: EvidenceBridgeKind? = nil, in bridges: [PromoteFlow.BridgeFact]) -> PromoteFlow.BridgeFact?` (end-set match, optional kind filter, first in order). `bridge(for:row:in:)` becomes a one-line call to it with the relationship's kind. `PromoteModel.reasonText`'s `.via` case calls it with `kind: nil`; delete the inline predicate. The page's "via … role" text must not change.
Check: `grep -n "Set(\[bridge.endA\|Set<String?>(\[\$0.endA" macos/App` prints only the one line inside the new helper; tests pass; scripts/check-pr.sh PASS.

**F-2 — Member boxes under a non-pinnable line are disabled** (from UX-1)
Files: `PromoteEvidenceDisplay.swift`, `PromoteEvidenceSheet.swift`, tests.
Failing test first: `memberBoxIsDisabledUnderANonPinnableLine` — `PromoteEvidenceDisplay.memberDisabled(line:)` (or similar) is true for a `pinnable == false` line and false for a pinnable one.
Do: `EvidenceMemberRow` takes `isDisabled: Bool` and passes it to `PVCheckbox(isDisabled:)`; both line views pass the helper's answer (the same rule the line box uses, `!line.pinnable`).
Check: test passes; check-pr PASS.

**F-3 — One line view and shared card pieces** (from MNT-2)
Files: `PromoteEvidenceSheet.swift`.
Do: replace `EvidenceLineRow` and `EvidenceLikenessLine` with one private `EvidenceLine` view that owns ids, key, isOpen, collapses, the disclosure, the checkbox, the Badge column and the expanded member list, with a `@ViewBuilder` middle-column slot and a `spoken: String` input. Add one dispatcher view (`EvidenceLineView(context:line:expanded:)`) that picks the likeness middle column for a `.depiction` line, and use it at both dispatch sites (property Card, relationship section). Add private `EvidenceCardHeader(title:)` for the sunken h4 header (section and Lifespan Cards) and `EvidenceOutcomeBadge(outcome:cause:)` for the four `PVBadge(L10n.Promote.outcome…)…frame(width: badgeColumn…).accessibilityHidden(true)` copies. Identifiers from D-17 (`promote.sheet.pin.*`, `promote.sheet.likeness.*`, `promote.sheet.expand.*`, `promote.sheet.member.*`) stay exactly as today.
Check: `grep -c "PVBadge(L10n.Promote.outcome" macos/App/Features/Promote/PromoteEvidenceSheet.swift` prints 1; `grep -n "struct EvidenceLineRow\|struct EvidenceLikenessLine" macos/App` prints nothing; check-pr PASS.

**F-4 — Display choices move to PromoteEvidenceDisplay, with tests; spoken labels survive an untitled Source** (from MNT-3, UX-3, outside-lens concatenation)
Files: `PromoteEvidenceDisplay.swift`, `PromoteEvidenceSheet.swift`, `macos/App/Platform/L10n.swift`, `macos/App/Resources/Localizable.xcstrings`, tests. Skill: `add-localized-string`.
Failing tests first, in `PromoteEvidenceDisplayTests`: `spokenLabelOfACollapsedLineCounts`, `spokenLabelOfOneMemberNamesItsSource`, `spokenLabelWithoutASourceTitleSpeaksTheValue`, `detailPartsOfOneMemberLine`, `linkPrefersTheBridgeElseTheNeighbor` (empty end name or phrase → "with <name>"; missing neighbor → `sheetNeighbor`), `nothingToCompareWhenNoLineRelationshipOrLifespan`, `likenessMemberWithoutTitleUsesFallback`.
Do: move into `PromoteEvidenceDisplay` as pure functions (taking the plain values the view has — labels, worded value, source title as `String?` — not the view context): the spoken label of a line (count form / line-plus-member form), the detail parts, the relationship link text (built on F-1's `bridge`), `hasNothingToCompare(_:)`, and the member-source / member-text fallbacks. Views call them. Replace `base + " " + spokenMember(...)` with one localized format key (e.g. `spokenLineMember` "%1$@ %2$@" style that takes the whole sentence and member, or a single key with all arguments) so translators can reorder. When the Source title is nil or empty, the spoken member is the value alone (new key if needed, e.g. `spokenMemberUntitled`, or reuse the value) — never ": value". For a likeness member with no title, use a fallback word (`sheetNeighbor`-style "a record" key — add `L10n.Promote.untitledRecord` "a record" if none exists), never the depiction's "Likeness". Remove keys this makes dead.
Check: tests pass; `grep -n '+ " " +' macos/App/Features/Promote/PromoteEvidenceSheet.swift` prints nothing; check-pr PASS.

**F-5 — No intro while loading or after a failed read** (from CON-1, UX-2)
Files: `PromoteEvidenceSheet.swift`.
Do: the intro and conflict callout show only when evidence is loaded and not empty (`if let evidence { if hasNothingToCompare … else intro … }`). Loading and Failed show no intro (D-17).
Check: check-pr PASS.

**F-6 — An expanded likeness member shows its crop** (from CON-2)
Files: `PromoteEvidenceSheet.swift`.
Do: in the member row, when `pin.value` is `.depiction`, draw `LikenessTile` (no overlay, `LikenessTile.crop`) in place of the worded value, Source title under it as today.
Check: check-pr PASS.

**F-7 — `lineSentence` switches on a value family** (from MNT-4)
Files: `macos/App/Platform/L10n.swift`.
Failing test first: none new; `lineSentencesFollowTheTable` and `badgesFollowTheOutcome` must stay green unchanged.
Do: private `enum ValueFamily { case name, date, likeness, other }` derived once from the value; switch on it in place of the three Bools; the disagreement wording returns from one place only; remove the unreachable `.different` arm.
Check: existing tests pass unchanged; check-pr PASS.

**F-8 — One crop size** (from MNT-5, CON-3)
Files: `PromoteEvidenceSheet.swift`.
Do: delete `EvidenceLayout.crop`; use `LikenessTile.crop`.
Check: `grep -n "EvidenceLayout.crop" macos/App` prints nothing.

**F-9 — DesignSystem README** (from MNT-6)
Files: `macos/App/DesignSystem/README.md`.
Do: Panel row adds "title accessory (right end of the title row)"; Checkbox row adds that `isIndeterminate` is spoken "partly checked". Also the Field row, if it names `required`, now names `marker: .required / .optional`.
Check: check-pr PASS.

## Round 1 — fix
Head `3f2451a9`. Gate PASS (orchestrator re-ran). F-1 – F-9 done in one commit (they share `PromoteEvidenceSheet.swift`); failing-first runs failed on compile (new API absent). New keys `spokenLineMember`, `spokenMemberUntitled`, `untitledRecord`. `lineSentence`'s `.different` now shares the `.unknown, .unrecognized` arm. `completed.md` paragraph gained a clause.
CON minors read by the orchestrator: CON-1 **Resolved** (intro inside `if let evidence`, `PromoteEvidenceSheet.swift:130-151`); CON-2 **Resolved** (member row draws `LikenessTile` for `.depiction`, `:567-575`); CON-3 Resolved by F-8.

## Round 2 — review
Reviewers: MNT, UX (each had a Fix-now major). Not sent: CON (minors only, read above), SEC+PRF (clean). Scope `ec0414a2...3f2451a9`.

| ID | Lens | Severity | Title | Decision | Evidence / reason |
| --- | --- | --- | --- | --- | --- |
| MNT-1 | MNT | major | (round 1) one bridge lookup | Resolved | `PromoteEvidenceDisplay.bridge(between:and:kind:in:)` is the only end-set match |
| MNT-2 | MNT | major | (round 1) line view near-copies | Resolved | one `EvidenceLine<Middle>`, `EvidenceLineView`, `EvidenceCardHeader`, `EvidenceOutcomeBadge` |
| MNT-3 | MNT | major | (round 1) display choices in views | Resolved | pure functions in `PromoteEvidenceDisplay`, tests `PromoteEvidenceDisplayTests.swift:318-380` |
| UX-1 | UX | major | (round 1) member boxes under non-pinnable line | Resolved | `memberDisabled(line:)` → `PVCheckbox(isDisabled:)`; test `memberBoxIsDisabledUnderANonPinnableLine` |

No new blockers. No Fix plan 2.

Nits (logged): `MNT-10 · "empty title means none" checked three times (EvidenceContext.sourceTitle, Member.title, likenessSource)` · `MNT-11 · PromoteEvidenceSheet.swift:449 · line box disable rule inline beside memberDisabled` · `MNT-12 · :557-565 · spokenSource still a display choice in the view` · `UX-7 · :487-494 · untitled likeness member spoken "a record: Likeness."` · `UX-8 · PromoteEvidenceDisplay.swift:114 · visible caption "a record" lowercase`

## Finish
Merged `origin/main` @ `1e8a5802` (docs only: S12-08 plan). Gate PASS on the merge `86e2c22e`.

## Result
Gate PASS · CI pending · open: none · follow-ups: — · to user: — (manual check §10 not run) · CON miss: none
