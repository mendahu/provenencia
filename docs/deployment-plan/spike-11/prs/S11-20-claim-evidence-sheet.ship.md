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

## Round 3 — user request (2026-10-22)
The user asked: the record subrows under an expanded line read on one line, with the value on the left and the Source aligned to the far right, so each subrow is shorter.

### Fix plan 3
**F-10 — Expanded record subrows on one line** (from the user)
Files: `macos/App/Features/Promote/PromoteEvidenceSheet.swift` (`EvidenceMemberRow`).
Failing test first: none (layout only; no view test target).
Do: in `EvidenceMemberRow.content`, change the non-depiction branch from a `VStack` (value over Source title) to one row: `HStack(alignment: .firstTextBaseline, spacing: PVSpacing.space3)` with the worded value leading (`bodySmall`, `textPrimary`, `lineLimit(1)`, `truncationMode(.tail)`), a `Spacer(minLength: PVSpacing.space3)`, and the Source title trailing (`caption`, `textMuted`, `lineLimit(1)`, `truncationMode(.middle)`, `multilineTextAlignment(.trailing)`). Give the value higher `layoutPriority` so a long Source title truncates first. If there is no title, the value stands alone. Change the row's vertical padding from `PVSpacing.space3` to `PVSpacing.space2`, and align the checkbox HStack `.center` for text members so the box sits on the line's middle. The depiction branch (likeness tile with its caption) keeps its layout and `.top` alignment. Spoken label, identifiers, and the disabled rule do not change.
Check: scripts/check-pr.sh PASS.

## Round 3 — fix
Head `1e948a9e`. Gate PASS (orchestrator re-ran). F-10 done as planned; diff read by the orchestrator (one file, `EvidenceMemberRow` only). Layout not checked by eye. No review round: a user-requested layout change, not a finding.

## Result (after round 3)
Gate PASS · CI see PR · open: none · follow-ups: — · to user: — (manual check §10 not run) · CON miss: none

## Round 4 — user request (2026-10-22)
The user asked: a relationship card's heading becomes the connection as one sentence ("Marion Smith is the mother of Don Smith"), with the outcome as a badge. When the page has no bridge for the relationship, keep today's two lines. This supersedes plan **D-14**'s header shape (outcome sentence over `connectionLink`) by the user's decision; record it in the plan's Landed note.

### Fix plan 4
**F-11 — Relationship heading is the bridge sentence plus an outcome badge** (from the user)
Files: `macos/App/Features/Promote/PromoteEvidenceDisplay.swift`, `macos/App/Features/Promote/PromoteEvidenceSheet.swift` (`EvidenceRelationshipCard.header`, `linkText`), `macos/ProvenenciaTests/PromoteEvidenceDisplayTests.swift`, the S11-20 plan's Landed note. Skill: `add-localized-string` only if a key changes.
Failing test first: `PromoteEvidenceDisplayTests` `relationshipHeaderIsTheBridgeSentence` — with a bridge whose `sentence` is "Marion Smith is the mother of Don Smith", the header's title is that sentence, it has no second line, it shows the badge, and its spoken label is "Marion Smith is the mother of Don Smith. Agrees." (`.partial` → "… Similar."); with `bridge: nil` the title is `relationshipSentence(outcome)` ("They share a matching connection"), the second line is "with <neighbor name>" (or "with a connected record" when the name is nil), no badge, and the spoken label is today's `spokenRelationship(sentence:link:)`.
Do: replace `PromoteEvidenceDisplay.linkText(bridge:neighborName:)` with `relationshipHeader(bridge: PromoteFlow.BridgeFact?, outcome: EvidenceOutcome, neighborName: String?) -> RelationshipHeader`, a small struct `{ title: String; subtitle: String?; showsBadge: Bool; spoken: String }`. A found bridge (from `bridge(for:row:in:)`) gives `title: bridge.sentence` (already built by `EvidenceBridgeEdgeSummary.sentence` in `PromoteModel.facts()`; never empty), `subtitle: nil`, `showsBadge: true`, `spoken: spokenRelationship(sentence: bridge.sentence, link: L10n.Promote.outcome(outcome))`. No bridge gives today's two lines and today's spoken label, no badge. In `EvidenceRelationshipCard.header`: title in the current medium `bodySmall` style; the subtitle row (`caption`, `textSecondary`) only when `subtitle != nil`; `· N records` stays on the subtitle row when there is one, and otherwise sits after the title in the title row (`caption`, `textMuted`). When `showsBadge`, put `EvidenceOutcomeBadge(outcome: relationship.outcome)` at the trailing end of the header row, so it lines up with the line badges below; align the header `HStack` `.center` when there is no subtitle, `.top` when there is. Checkbox, its identifier, its disabled rule and the card body do not change. Update the existing `linkPrefersTheBridgeElseTheNeighbor` test to the new helper (or fold it into the new test) and drop `linkText`. Keep `L10n.Promote.connectionLink` and its test: S11-27 is planned to call it. Keep `relationshipAgrees` / `relationshipSimilar` (fallback). Add one sentence to the plan's Landed note: "D-14's header shape was changed at the user's request in ship round 4: the bridge sentence with an outcome badge; the outcome sentence over 'with <name>' remains the fallback."
Check: the new test passes; `grep -n "linkText" macos/App macos/ProvenenciaTests -r` prints nothing; scripts/check-pr.sh PASS.

## Round 4 — fix
Head `744ed015`. Gate PASS (orchestrator re-ran). F-11 done; diff read by the orchestrator. One miss in the fix plan itself: the pre-change header always drew `EvidenceOutcomeBadge`, so "no badge" in the fallback changed the fallback, which the user asked to keep as it was. → F-12.

**F-12 — The fallback keeps its badge** (orchestrator, against the user's "keep the two-line fallback")
Files: `PromoteEvidenceDisplay.swift`, `PromoteEvidenceSheet.swift`, `PromoteEvidenceDisplayTests.swift`.
Do: the badge shows in both cases, as it did before round 4. Drop `RelationshipHeader.showsBadge` and the `if heading.showsBadge` guard; the header always ends with `EvidenceOutcomeBadge(outcome: relationship.outcome)`. Update `relationshipHeaderIsTheBridgeSentence` to drop the `showsBadge` expectations. Everything else in F-11 stays.
Check: test passes; `grep -n showsBadge -r macos` prints nothing; scripts/check-pr.sh PASS.
F-12 done at `da4f3ffe`; gate PASS (orchestrator re-ran); diff read: `showsBadge` gone, the badge always ends the header. No review round: user-requested copy and layout, not a finding.

## Result (after round 4)
Gate PASS · CI see PR · open: none · follow-ups: — · to user: — (manual check §10 not run) · CON miss: none

## Round 5 — user request (2026-10-22)
The user asked: person-to-event connections read in plain language for birth, death, marriage, baptism and funeral (and burial, approved), and every connection sentence comes from a registry, not `if`/`switch` branches per kind. Wording approved by the user:

| Event | Role | Sentence (en) |
| --- | --- | --- |
| birth | subject | Don Smith was born |
| death | subject | Don Smith died |
| baptism | subject | Don Smith was baptized |
| marriage | subject | Don Smith married (Marriage of Don Smith and Marion Brown) |
| funeral | subject | Don Smith's funeral (Funeral of Don Smith) |
| burial | subject | Don Smith was buried |
| any | any other | unchanged: "%1$@ participated as %2$@ at %3$@" |

### Fix plan 5
**F-13 — Connection-sentence registry** (from the user)
Files: `macos/App/Features/Catalog/Connect/ConnectionSentenceRegistry.swift` *(new)*, `macos/App/Features/Catalog/Connect/EvidenceBridgeEdgeSummary.swift`, `macos/App/Features/Catalog/Connect/ConnectEndpointBinding.swift`, `macos/App/Platform/L10n.swift`, `macos/App/Resources/Localizable.xcstrings`, `macos/ProvenenciaTests/ConnectionSentenceRegistryTests.swift` *(new)*, `macos/ProvenenciaTests/EvidenceBridgeEdgeSummaryTests.swift`, `macos/Provenencia.xcodeproj/project.pbxproj`. Skills: `add-localized-string`, `add-swift-test`.

Failing tests first, in `ConnectionSentenceRegistryTests`:
- `eachEventRowReadsPlainly` — for each of the six approved rows (participation, role key `subject`, event type key `birth|death|baptism|marriage|funeral|burial`), with person "Don Smith" and event "Marriage of Don Smith and Marion Brown" / "Funeral of Don Smith" etc., the sentence equals the table above.
- `anotherRoleKeepsTheGenericSentence` — `witness` at `baptism` → "John Smith participated as Witness at Baptism of Don Smith" (role label as `PropertyTermDisplay` gives it today).
- `lookupPrefersTheMostSpecificRow` — event type + role row beats a role-only row beats the kind default; an unknown event type key falls to the default.
- `everyKindHasADefaultRow` — for every `EvidenceBridgeKind.allCases` (add `CaseIterable` if missing), the registry has a row with no term and no event type, and it yields today's sentence, today's no-term fallback, and today's phrase forms (`bridgeSummary…TypeOnly` / `…Bare`).
- In `EvidenceBridgeEdgeSummaryTests`: `aParticipationReadsFromTheRegistry` — a placed participation bridge whose role observation has term key `subject` and whose event endpoint has an `event_type` observation with term key `birth` reads "Alice was born".
- `anEventEndpointIsNamedByItsTitle` — an event endpoint with an `eventTitle` is named by `EventTitleDisplay.title`, not by its `event_type` value.

Do:
1. **Registry** (`ConnectionSentenceRegistry`): a static array of `Row { kind: EvidenceBridgeKind; term: String?; eventType: String?; sentence: (Slots) -> String; noTerm: ((Slots) -> String)?; phrase: ((String?) -> String)? }` where `Slots { a: String; b: String; termLabel: String? }` are the two endpoint nouns in the kind's A/B order (the order `ConnectEndpointBinding` documents: participation A = person, B = event; location A = event, B = place; relationship and place relationship A = from, B = to) and the term's display label. One `lookup(kind:term:eventType:) -> Row` picks the first row matching kind + term + eventType, then kind + term, then the kind's default row (term and eventType nil). The table is data: no per-kind `switch` or per-key `if` outside it. Rows: the four kind defaults, built from today's L10n keys (`bridgeSummaryRelationship` / `…Fallback` / `…TypeOnly` / `…Bare`, and the same for location, place relationship, participation), and the six approved participation rows, each with a new L10n key under `L10n.EvidenceGraph` (`bridgeSentenceBirthSubject` "%1$@ was born", `…DeathSubject` "%1$@ died", `…BaptismSubject` "%1$@ was baptized", `…MarriageSubject` "%1$@ married (%2$@)", `…FuneralSubject` "%1$@'s funeral (%2$@)", `…BurialSubject` "%1$@ was buried"; comments name the arguments). A specific row with no `phrase` uses its kind default's `phrase` and `noTerm`.
2. **`EvidenceBridgeEdgeSummary`**: `sentence(kind:person:related:event:place:term:)`, `phrase(kind:term:)` and the private per-kind `…Sentence` helpers go; `sentence(for:in:orders:)` and `phrase(for:)` read the registry. Read the role / relationship type term **key** (`observation.valueTermKey`) beside the label `termDisplay` already makes, and the event endpoint's `event_type` term key from the placed event's observations (positive first, as `identityNoun` picks). Keep the never-empty and whole-name fallbacks exactly as they are.
3. **Event nouns**: in `noun(subjectID:snapshot:orders:)`, an event endpoint with an `eventTitle` is named by `EventTitleDisplay.title(_, orders:)` (as `SourceGraphPlacedSubject.displayName` does); otherwise the existing identity / label / type-and-ref chain.
4. **`ConnectEndpointBinding.sentence(kind:endpointA:endpointB:term:)`**: drop its per-kind `switch`; it calls the registry with A/B as given (no term key, no event type → the default row).
5. Remove L10n keys left with no reader; keep every key a default row uses.

Check: new and updated tests pass; `grep -nE "case \.(location|relationship|placeRelationship|participation)" macos/App/Features/Catalog/Connect/EvidenceBridgeEdgeSummary.swift macos/App/Features/Catalog/Connect/ConnectEndpointBinding.swift` prints only lines inside the registry's row table, or nothing (the `termDisplay` property-key lookup may become a per-kind field on the default row instead; either way no `switch kind` remains in those two files); scripts/check-pr.sh PASS. Add one sentence to the plan's Landed note: "Ship round 5, at the user's request: connection sentences come from `ConnectionSentenceRegistry`, with plain wording for a subject's birth, death, baptism, marriage, funeral and burial, and an Event endpoint named by its title."

## Round 5 — fix
Head `d4866a28` (`b471c537` F-13; `d4866a28` Landed note, which also swept in the uncommitted ship log via `git commit -a`). Gate PASS (orchestrator re-ran). Failing-first not run: the registry type did not exist.

## Round 5 — review
Reviewers: MNT, CON (Sonnet), scope `fdbf0637...d4866a28`. The user asked for this change; the review checks it against Fix plan 5 and the user's "registry-driven, no per-kind if" requirement. Not sent: UX (copy approved by the user; no new chrome), SEC+PRF (no paths matched).

| ID | Lens | Severity | Title | Decision | Evidence / reason |
| --- | --- | --- | --- | --- | --- |
| MNT-R5-1, CON-3 | MNT, CON | blocker | Registry spells seeded Property and term keys as raw strings, outside `ConclusionVocabulary.swift` and its Go drift test | Fix now → F-14 | verified: `.cursor/rules/macos-client.mdc:20`; `SeededPropertyKey` lacks `role`/`event_type`/…; `swift_vocabulary_test.go` "seeded term keys are seeded" covers only gender and sex_at_birth |
| MNT-R5-2 | MNT | major | `bridge.kind == .participation ? endpointB : nil` is a per-kind branch outside the table | Fix now → F-15 | verified `EvidenceBridgeEdgeSummary.swift:43-47`; against the user's requirement |
| MNT-R5-3 | MNT | minor | `phrase`/`noTerm`/`termProperty` read only from the default row; a specific row's `phrase` is silently ignored | Fix now → F-15 | code this PR adds |
| MNT-R5-4 | MNT | minor | `ConnectEndpointBinding.sentence(kind:…)` only forwards | Fix now → F-15 | one caller, same file |
| MNT-R5-5 | MNT | minor | No uniqueness check on (kind, term, eventType) | Fix now → F-15 | |
| MNT (outside), CON-1 | CON | minor | Role-only lookup step untested; plain wording gated on the label, not the key | Fix now → F-15 | `ConnectionSentenceRegistry.swift:104`, `:135` |
| CON-2 | CON | minor | No end-to-end test with both an `eventTitle` and an `event_type` key (marriage, funeral) | Fix now → F-15 | |
| — | MNT | same-pass | `L10n.PropertyTerm.resource` switch spells the same seeded term keys as literals | Fix now → F-16 | same-pass cleanup: this PR adds copies of those keys |

Nits (logged, riding along in F-15): `CON-4 · Localizable.xcstrings · six bridge.sentence comments word arguments differently from L10n.swift` · `CON-5 · ConnectionSentenceRegistry.swift:99 · force unwrap in defaultRow` · `CON-6 · split everyKindHasADefaultRow` · `MNT-R5-6 · :120 · ?? "role" hides a missing termProperty` · `MNT-R5-7 · :137 · fallback computed before the early return`

### Fix plan 6
**F-14 — Seeded keys spelled once, checked against Go** (from MNT-R5-1, CON-3)
Files: `macos/App/Platform/ConclusionVocabulary.swift`, `api/ffi/swift_vocabulary_test.go`, `macos/App/Features/Catalog/Connect/ConnectionSentenceRegistry.swift`, `macos/App/Features/Catalog/Connect/EvidenceBridgeEdgeSummary.swift`.
Failing test first: add `role`, `relationship_type`, `place_relationship_type`, `event_type`, `person`, `event`, `place` (only those the registry or `EvidenceBridgeEdgeSummary` actually spell; check each is seeded with `go test ./api/ffi -run TestSwift`) to `SeededPropertyKey`; add `SeededRoleTerm` and `SeededEventTypeTerm` enums; add `{"SeededRoleTerm","role"}` and `{"SeededEventTypeTerm","event_type"}` to the Go "seeded term keys are seeded" table, and run it (it must pass, proving the keys are seeded; a misspelled key must fail it — try one, then revert).
Do: the registry rows, `termProperty`, and `EvidenceBridgeEdgeSummary.identityPropertyKey` use these constants; no seeded key is a string literal in either file.
Check: `grep -nE '"(role|relationship_type|place_relationship_type|event_type|birth|death|baptism|marriage|funeral|burial|subject)"' macos/App/Features/Catalog/Connect` prints nothing; `go test ./api/ffi/...` PASS.

**F-15 — The table owns every per-kind fact** (from MNT-R5-2 – 5, CON-1, CON-2, nits)
Files: `ConnectionSentenceRegistry.swift`, `EvidenceBridgeEdgeSummary.swift`, `ConnectEndpointBinding.swift`, `macos/ProvenenciaTests/ConnectionSentenceRegistryTests.swift`, `macos/ProvenenciaTests/EvidenceBridgeEdgeSummaryTests.swift`, `Localizable.xcstrings`.
Failing tests first: `aSpecificRowsPhraseWins` (a test-only row with its own `phrase` is used by `phrase(kind:term:…)`); `rowsAreUnique` (no two rows share kind, term, eventType); `aRoleOnlyRowBeatsTheDefault` (exercise the middle lookup step, with a row injected through a test seam such as `lookup(…, in rows:)`); `aKeyWithoutALabelStillReadsPlainly` (termKey `subject`, label nil → "Don Smith was born"); in `EvidenceBridgeEdgeSummaryTests`, `aMarriageReadsWithItsTitle` and `aFuneralReadsWithItsTitle` (placed snapshot with an event that has both an `eventTitle` and an `event_type` observation) → "Don Smith married (Marriage of Don Smith and Marion Brown)" / "Don Smith's funeral (Funeral of Don Smith)".
Do:
- Split `Row` into a per-kind `KindEntry { kind; termProperty (required); sentence; noTerm; phrase }` list and a `Sentence { kind; term; eventType; sentence; phrase? }` list of specific rows; `lookup` returns the specific sentence if any, else the kind entry's; `phrase` uses the specific row's `phrase` when set, else the kind entry's. No `?? "role"` default; no force-unwrap (a missing kind entry is a `preconditionFailure` with a message; `everyKindHasAnEntry` proves none is missing).
- Plain wording keys on the term **key**, not the label: a matching specific row needs the key; the label only fills `%@` slots.
- Remove the kind test from `EvidenceBridgeEdgeSummary`: the event type comes from whichever endpoint is an event (`eventTypeKey(A) ?? eventTypeKey(B)`).
- Inline `ConnectEndpointBinding.sentence(kind:endpointA:endpointB:term:)` into `pendingSentence` (calls the registry) and delete the wrapper.
- Move `let fallback` below the early return; align the six xcstrings comments with `L10n.swift`; split `everyKindHasADefaultRow` into existence and parity tests.
Check: tests pass; `grep -n "\.participation\|\.location\|\.relationship\|\.placeRelationship" macos/App/Features/Catalog/Connect/EvidenceBridgeEdgeSummary.swift macos/App/Features/Catalog/Connect/ConnectEndpointBinding.swift` prints nothing; scripts/check-pr.sh PASS.

**F-16 — `L10n.PropertyTerm` reads the typed keys** (same-pass cleanup)
Files: `macos/App/Platform/L10n.swift`, `macos/App/Platform/ConclusionVocabulary.swift`, `api/ffi/swift_vocabulary_test.go`.
Do: every term key in `L10n.PropertyTerm.resource`'s switch (`event_type`, `role`, `relationship_type`, `place_relationship_type`) comes from a `Seeded…Term` enum (`SeededEventTypeTerm` and `SeededRoleTerm` from F-14 grow to the full seeded lists; add `SeededRelationshipTypeTerm` and `SeededPlaceRelationshipTypeTerm`), and the property keys from `SeededPropertyKey`. Add the two new enums to the Go "seeded term keys are seeded" table. The strings the switch returns do not change.
Check: `grep -nE 'case \("' macos/App/Platform/L10n.swift` prints nothing inside `PropertyTerm`; `go test ./api/ffi/...` PASS; scripts/check-pr.sh PASS.

## Round 6 — fix
Head `6b324baf` (`69a2e148` F-14, `77aa1576` F-15, `6b324baf` F-16). Gate PASS (orchestrator re-ran). Misspelled-key trial (`funerall`) failed the Go drift test, then reverted. `?? .participation` for an unknown bridge type moved into `ConnectionSentenceRegistry.kind(bridgeTypeKey:)`. Failing-first not run for F-15 (new API).

## Round 6 — review
Reviewers: MNT (had the blocker). CON minors read by the orchestrator against the F-15 tests (`aKeyWithoutALabelStillReadsPlainly`, `aMarriageReadsWithItsTitle`, `aFuneralReadsWithItsTitle`, `aRoleOnlyRowBeatsTheDefault`): CON-1, CON-2 **Resolved**; CON-3 Resolved by F-14. Scope `d4866a28...6b324baf`.
MNT round 2: MNT-R5-1 – MNT-R5-5 all **Resolved**; no new blockers, no nits. Noted (not a finding): `phrase(for:)` passes no event type, so a future event-specific `Sentence` with its own `phrase` would not reach the card phrase; no row sets one today.

## Result (after round 6)
Gate PASS · `main` unchanged since the last merge · CI see PR · open: none · follow-ups: — · to user: — (manual check §10 not run) · CON miss: none
