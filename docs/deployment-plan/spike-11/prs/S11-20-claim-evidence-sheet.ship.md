# Ship log — S11-20 Claim evidence sheet

**Plan:** [`S11-20-claim-evidence-sheet.md`](S11-20-claim-evidence-sheet.md) · **Branch:** `s11-20-claim-evidence-sheet` · **PR:** #525
**State:** fixing 8 (user report)

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

## Round 7 — user report (2026-10-22)
The user saw "Marion Robins participated as Subject at Death of Marion Robins" in the built app: the death row did not match.
Root cause (orchestrator, verified): the engine never sends an Observation's term **key**. `engine.proto` `Observation` has `value_term_id` but no key; `GoStore.mapObservation` (`GoStore.swift:2336`) leaves `CatalogObservation.valueTermKey` at `""`; only `FakeStore` fills it (`FakeStore.swift:3368`), so every Swift test passed. The registry matches on term keys, so in the real app no specific row ever matches. The same gap means `PropertyTermDisplay` has never used its L10n term labels in the real app (hence "Subject" from the catalog label). `CON miss: none` (no reviewer could see it from the Swift tests); recorded as a gap in the test seam.

### Fix plan 7
**F-17 — The engine sends the term key with every listed Observation** (from the user's report)
Files: `api/proto/engine.proto`, generated `api/proto/engine/*.pb.go` and `macos/App/Platform/Generated/engine.pb.swift` (via `scripts/generate-proto.sh`), `core/database/observations/observations.go`, `api/ffi/handlers/observations.go`, `macos/App/Platform/GoStore.swift`, Go and Swift tests.
Failing tests first:
- Go: in the observations package tests, a listed term Observation (seeded `event_type` = `death`, and `role` = `subject`) has `ValueTermKey == "death"` / `"subject"`; a non-term Observation has `""`.
- Go: in the handler tests (`runRPC`), the RPC that lists a Source's graph observations (the one feeding `SourceGraphSnapshot`) returns `value_term_key` for a term Observation.
- Swift: `GoStore.mapObservation` carries `valueTermKey` from the proto (a test that builds a `Provenencia_Engine_V1_Observation` with `valueTermKey = "death"`).
Do: add `string value_term_key = 20;` to `Observation` (response data, never read from a request; comment it like `depiction`). Select `property_terms.key` beside the label the `Listed` query already joins, as `Listed.ValueTermKey`. `listedObservationProto` sets `ValueTermKey`. `GoStore.mapObservation` maps it. Regenerate the proto code with the repo script. Do not change `FakeStore`.
Check: tests pass; scripts/check-pr.sh PASS. In the Landed note, add: "Ship round 7: the engine now sends an Observation's term key, which the registry and `PropertyTermDisplay` match on."

## Round 7 — fix
Head `3f24103f`. Gate PASS (orchestrator re-ran). F-17 done: `Listed.ValueTermKey` (`COALESCE(pt.key,'')` beside the joined label), `listedFromSource` in `source_read.go` also sets it (the graph read builds rows there; not in the file list, needed), proto field `value_term_key = 20`, `GoStore.mapObservation`. `ListObservationsBySource` feeds `CatalogQueryRegistry.swift:245`, the snapshot behind Promote and the Evidence graph. Tests: observations package (`female`), handler `TestListObservationsBySource` (`birth`, failed `""` with the fix stashed), Swift `mapsTheTermKey`.
No reviewer dispatched: `api/ffi/**` and `core/database/**` paths matched SEC+PRF, but the diff adds one response-only string from a column the query already joins (no new input, no new loop or query); read by the orchestrator.

## Result (after round 7)
Gate PASS · CI see PR · open: none · follow-ups: — · to user: — (manual check §10 not run) · CON miss: none (FakeStore filled a field the real store never sent; no Swift test could see it)

## Round 8 — user report (2026-10-22)
The user saw a lifespan **Conflicts** for "bef 11 Jan 2021 here, 27 Jul 1932 – 31 Mar 2001 on the record". A death before 11 Jan 2021 fits a 2001 death.
Root cause (orchestrator, verified in `core/subjectprofile/subjectprofile.go` `lifespan`): each date's open end is collapsed onto its closed end (`if first.Open { first = last }`). A lone `BEF 2021-01-11` death becomes the point 11 Jan 2021, so this side's span is that day, which falls outside the record's window (1932 – 2001) → conflict. The collapse is right for anchoring the window (`ShiftYears` needs a closed end) but wrong for the span and for the "date wholly outside the window" check, where an open end means the event can lie anywhere on that side. The code is S11-32's; the same builder feeds matching (`subjectcompare`), so the fix also corrects matching scores for open-ended dates.

### Fix plan 8
**F-18 — An open-ended date keeps its open end in the span** (from the user's report)
Files: `core/subjectprofile/subjectprofile.go` (`lifespan`), `core/subjectprofile/lifespan_test.go`; `core/subjectcompare` tests if the end-to-end case sits better there. Skill: `add-match-rule` (read its rules for the layer; no new knob).
Failing tests first (table in `lifespan_test.go`, plus one end-to-end through `subjectcompare` so the outcome, not just the span, is checked):
- `aDeathBeforeALaterDayFitsAnEarlierDeath` — side A: death `BEF 2021-01-11` only; side B: birth 1932-07-27, death 2001-03-31 → the lifespan outcome is not conflict (the user's case).
- `aDeathBeforeTheOtherBirthConflicts` — A: death `BEF 1931`; B: birth 1932, death 2001 → conflict.
- `aBirthAfterTheOtherDeathConflicts` — A: birth `AFT 2002`; B: 1932 – 2001 → conflict.
- `aBirthAfterAnEarlierYearFits` — A: birth `AFT 1900`; B: 1932 – 2001 → not conflict.
- Existing cases (points, ranges, a date wholly outside the window, undated) keep their results.
Do: in `lifespan`, keep each date's raw ends beside the collapsed ones. The collapsed ends still anchor the window (`spanFirst`/`spanLast` for `ShiftYears`, unchanged). The span's first end is the window's `start` when any date's raw first end is open, else `Earliest(early)`; its last end is the window's `end` when any raw last end is open, else `Latest(late)` (the events lie in the window, so the window bounds an open span). The "wholly outside the window" check uses raw ends and skips an open one (`(!rawLate.Open && rawLate.Hi < start.Lo) || (!rawEarly.Open && rawEarly.Lo > end.Hi)`). `side.First`/`side.Last` stay the dates as written. Update the function's doc comment to say how an open end is read.
Check: new tests pass and failed first; `go test -tags fts5 ./core/...` PASS; scripts/check-pr.sh PASS. Add to the plan's Landed note: "Ship round 8: an open-ended date (BEF/AFT) keeps its open end in a lifespan's span, so 'death before 2021' fits a 2001 death."

## Round 8 — fix
Head `f7b92ee3`. Gate PASS (orchestrator re-ran). F-18 done as planned; conflict cases moved to `BEF 1900` / `AFT 2010` (Tolerance 2 reads a one-year gap as Near); AC-2.9's span string became "1730 and 1840".
Orchestrator, reading the diff: **F-18's rule was wrong (the plan's error, not the implementer's).** Widening an open end to the window's edge loses a closed date's bound. Example: baptism 1900 + death `BEF 1980` against a record 1850 – 1870 now spans ~1790 – 1980, overlaps 1850 – 1870, and no longer conflicts, though the 1900 baptism makes it impossible. The date comparer already reads open ends natively (`EndsDate` makes `BEF`/`AFT`; `Earliest`/`Latest` skip open ends), and a lifespan conflict is "no overlap", so the span should be built from the raw ends: its first end is the earliest *closed* first end, open when none is closed; its last end likewise. → F-19.

**F-19 — The span reads raw ends, open only where no date closes it** (corrects F-18)
Files: `core/subjectprofile/subjectprofile.go`, `core/subjectprofile/entity_test.go`, `core/subjectprofile/lifespan_test.go`.
Failing test first: `aClosedDateStillBoundsAnOpenSpan` in `TestEntityCompare` — side A: baptism 1900, death `BEF 1980`; side B: birth 1850, death 1870 → conflict (fails at `f7b92ee3`).
Do: drop the `openFirst`/`openLast` widening to `start`/`end`. The span is `EndsDate(firstEnd, lastEnd)` where `firstEnd = Earliest(rawEarly)` (an open `DateEnd` when not ok) and `lastEnd = Latest(rawLate)` (open when not ok). Every date has at least one closed end, so at least one of the two is closed; if `EndsDate` still fails (closed first after closed last), keep today's contradicts. The window keeps the collapsed anchors, and the raw-end "wholly outside" check stays as F-18 wrote it. `side.First`/`side.Last` unchanged. Update AC-2.9's expected span to what the open span now prints (a `BEF 1840` span), and the doc comment.
Check: all four F-18 cases and the new case pass; `go test -tags fts5 ./core/...` PASS; scripts/check-pr.sh PASS.
F-19 done at `ee53202a`; gate PASS (orchestrator re-ran); `aClosedDateStillBoundsAnOpenSpan` failed first (span 1790–1980, agree) and passes.

## Round 8 — review
Verifier (Opus), asked to disprove the open-end rule at `ee53202a` with concrete dates against the base `fe5b79b3`. **Refuted**, one regression:

| ID | Lens | Severity | Title | Decision | Evidence / reason |
| --- | --- | --- | --- | --- | --- |
| VER-1 | CON | blocker | Birth `BEF 1900` + death `AFT 1950` now "contradicts itself" | Fix now → F-20 | verified by the verifier's throwaway test at both commits: base span 1900–1950, agree; `ee53202a` `Earliest(rawEarly)`=1950 > `Latest(rawLate)`=1900, `EndsDate` fails → contradicts. Any BEF date earlier than any AFT date triggers it |
| VER-2 | CON | major | A lone open date's **window** is still anchored on its collapsed end: death `BEF 1840` alone vs 1600–1650, birth `AFT 1800` alone vs 1950–2000, birth 1900 + death `AFT 1950` vs a 1990 residence all conflict though they can be one life | To user | not new (same at base, S11-32); the same family as the user's report, but changing how a window bounds a life (MaxYears from an open date) is a matching-rule call |

Also verified: no pair that cannot be one life lost its conflict; undated sides unchanged; the span is read only by `subjectcompare.lifespan`; `First`/`Last` unchanged.

**F-20 — An inverted closed span falls back to the collapsed ends** (from VER-1)
Files: `core/subjectprofile/subjectprofile.go`, `core/subjectprofile/entity_test.go`.
Failing tests first, in `TestEntityCompare`: `aBirthBeforeAndADeathAfterStillFit` — A: birth `BEF 1900`, death `AFT 1950`; B: birth 1890, death 1960 → not conflict, and A is not contradicts; `twoOpenResidencesStillFit` — A: residence `AFT 1950`, residence `BEF 1900`; same B → not conflict.
Do: when both `firstEnd` and `lastEnd` are closed and `firstEnd.Lo > lastEnd.Hi`, build the span from the collapsed `firstClosed`/`lastClosed` (the base's span) instead of returning contradicts. Say so in the doc comment.
Check: new tests fail first and pass; all earlier F-18/F-19 cases pass; scripts/check-pr.sh PASS.

## Round 8 — fix 2
Head `1885e198`. Gate PASS (orchestrator re-ran). F-20 done: `aBirthBeforeAndADeathAfterStillFit` and `twoOpenResidencesStillFit` failed first at `ee53202a` and pass; F-18/F-19 cases pass. VER-1 **Resolved**. Implementer slip, checked: a stray `git stash pop` touched the unrelated s12-03 stash; working tree verified clean apart from this log, both stashes still listed with their files.
VER-2 (open dates still anchor the lifetime window) is **To user**.

## Result (after round 8)
Gate PASS · CI see PR · open: none · to user: VER-2, term origin (a user term keyed like a product term reads as the product term) · manual check §10 not run

## Round 9 — user decision (2026-10-22)
VER-2, answered by the user: "a lifetime anchored only by an open-ended date stays open on that side." Plan it with a test matrix and verify it the same way.

### The rule (orchestrator)
The window is the possible lifetime. Every event lies inside it, and a life is at most `MaxYears` long. Each side of the window takes its bound from **closed** ends only; an open end never pins it:
- **Start:** (1) the earliest closed first end among `Starts` events (birth); else (2) the earliest closed first end among all dates, shifted back `MaxYears`; else (3) open.
- **End:** (1) the latest closed last end among `Ends` events (death, burial…); else (2) the latest closed last end among all dates, shifted forward `MaxYears`; else (3) open.

Why each step is a true bound:
- (1) A closed birth bounds the start; a closed death bounds the end.
- (2) The death falls on or after every event, so it falls on or after any closed first end `x`; the birth then falls on or after `x − MaxYears`. The end side is symmetric.
- Step (2) no longer shifts from a `BEF` date's collapsed day. Example: baptism 1900 plus residence `BEF 1850` now starts at 1790, not 1740. That is tighter and still true.

**Never open on both sides:** an open start needs every date's first end open, and an open end needs every date's last end open. Every date has at least one closed end.

The "wholly outside the window" check skips an open window end. Keep "contradicts" when a closed start falls after a closed end. Keep "undated" when a shift cannot be placed.

The user asked for "open on that side". Step (2) still applies the 110-year cap when any exact date exists; a side is fully open only when no exact date bounds it. That keeps the registry's longest-life rule (`MaxYears`, an existing knob) rather than dropping it. The final report must say so.

### Fix plan 9
**F-21 — The lifetime window reads closed ends only** (from VER-2, user decision)
Files: `core/subjectprofile/subjectprofile.go` (`lifespan`, doc comment), `core/subjectprofile/entity_test.go`, `core/subjectprofile/lifespan_test.go`. Skill: `add-match-rule` (no new knob).
Failing tests first: a table in `TestEntityCompare`, side B as given, outcome of the lifespan rule.

| # | Side A | Side B | Want | At `1885e198` |
| --- | --- | --- | --- | --- |
| W1 | death `BEF 1840` only | birth 1600, death 1650 | not conflict | conflict |
| W2 | birth `AFT 1800` only | birth 1950, death 2000 | not conflict | conflict |
| W3 | birth 1900, death `AFT 1950` | residence 1990 only | not conflict | conflict |
| W4 | birth `BEF 1900`, death `AFT 1950` | birth 1890, death 1960 | not conflict | not conflict |
| W5 | death `BEF 1840` only | residence 1900 only | conflict | conflict |
| W6 | birth `AFT 1800` only | death 1790 only | conflict | conflict |
| W7 | birth 1900, death `AFT 1950` | residence 2020 only | conflict (past 1900 + 110) | conflict |
| W8 | baptism 1900, residence `BEF 1850` | birth 1700, death 1785 | conflict (A's start ≥ 1790) | conflict |
| W9 | birth `BEF 1900` only | death 1700 only | conflict (B died before A's latest possible birth?), see note | — |

Note on W9: A's window is open at the start and runs to 1900 + 110 at the end. B's window comes from death 1700, so it starts at 1590 and ends at 1700. B's span is 1700. A's window (−∞, 2010] contains 1700, so A's window direction fits. B's window [1590, 1700] against A's span `BEF 1900` overlaps. So W9 is **not conflict**, and that is correct: a birth before 1900 can be 1650. Write it as "not conflict".

W1–W3 must fail first. W4–W8 are regression guards; run them before the change and say which already pass.
Do:
- Implement the rule above.
- Keep the collapsed `early`/`late` only for `side.First`/`side.Last` and the inverted-span fallback (F-20).
- The window uses the raw ends.
- The "wholly outside" check skips an open window end as well as an open date end.
- List every existing test whose expected window or result changes, with one line on why the new value is right. Expect changes only where a window was anchored on a collapsed open date.

Check:
- the matrix passes;
- `go test -tags fts5 ./core/...` PASS;
- scripts/check-pr.sh PASS.

Landed note, add: "Ship round 9, at the user's request: a lifetime window takes its bounds from closed dates only; a side anchored only by an open-ended date stays open (the 110-year cap still applies from any exact date)."

## Round 9 — fix
Head `063c2e60`. Gate PASS (orchestrator re-ran). W1–W3 and W9 failed first; W4–W8 already passed. One existing case changed: AC-2.9 window "Between 1730 and 1840" → "Before 1840".

## Round 9 — review
Verifier (Opus): **Refuted, partly.** Questions (1), (4) and (5) were confirmed: no new false conflict; the window is never open on both sides; an open side contains anything. One defect:

| ID | Lens | Severity | Title | Decision | Evidence / reason |
| --- | --- | --- | --- | --- | --- |
| VER-3 | CON | blocker | A side with several Starts/Ends records, only some open, now "contradicts itself", and the conflicts that side should raise are lost | Fix now → F-22 | verified by the verifier's probe at both commits. C1 birth 1900 + birth `BEF 1895` vs death 1700: before, conflict; now A contradicts, unknown. C2–C5 are the same pattern, including death 1 Mar 1850 + burial `AFT 5 Mar 1850`. The exact-date control C6 (two exact births) still conflicts. Step (1) is a true bound only when every Starts (Ends) record is closed on that side |

Also confirmed rightly changed:
- "baptism 1900, residence `BEF 1780`" now contradicts (a life of more than 110 years).
- "birth `BEF 1900`, residence 1880" no longer contradicts.
- "death `AFT 1950`, residence 1980" no longer contradicts.

**F-22 — A record's bound counts only when every such record is closed** (from VER-3)
Files: `core/subjectprofile/subjectprofile.go`, `core/subjectprofile/entity_test.go`.
Failing tests first (side A vs side B → lifespan outcome; A must have a lifespan, not contradicts):

| # | Side A | Side B | Want |
| --- | --- | --- | --- |
| C1 | birth 1900, birth `BEF 1895` | death 1700 | conflict |
| C2 | death 1850, burial `AFT 1900` | birth 1960 | conflict |
| C3 | death 1 Mar 1850, burial `AFT 5 Mar 1850` | birth 1960 | conflict |
| C4 | birth `AFT 1890`, birth `BEF 1880` | residence 1700 | conflict |
| C5 | death 1850, death `AFT 1855` | birth 1960 | conflict |

Regression guards, already right at `063c2e60`:

| # | Side A | Side B | Want |
| --- | --- | --- | --- |
| C6 | birth 1900, birth 1895 | death 1700 | conflict |
| G1 | baptism 1900, residence `BEF 1780` | — | A contradicts |
| G2 | birth `BEF 1900`, residence 1880 | — | A has a lifespan |
| G3 | death `AFT 1950`, residence 1980 | — | A has a lifespan |

W1–W9 and every F-18 – F-20 case still pass.
Do: step (1) for the start applies only when every Starts record's first end is closed; otherwise step (2). The end works the same way with Ends records and last ends. Update the doc comment and the Round 9 rule text in the plan's Landed note if it states step (1) without this condition.
Check: tables pass; `go test -tags fts5 ./core/...` PASS; scripts/check-pr.sh PASS.
F-22 done at `b13fddbb`; gate PASS (orchestrator re-ran); C1–C5 failed first and pass; C6, G1–G3 pass.

## Round 9 — review 2
Verifier (Opus), probing against `fe5b79b3`, `1885e198` and `b13fddbb`. **Refuted, partly.** No new false conflict was found, and the self-contradiction changes are right. Two regressions remain against the base:

| ID | Lens | Severity | Title | Decision | Evidence / reason |
| --- | --- | --- | --- | --- | --- |
| VER-4 | CON | blocker | Step (2) drops the cross bound, so a closed death no longer caps the birth from below and a closed birth no longer caps the death from above | To user | Probes: P19 birth `BEF 1850`, residence 1850, death 1900 \| residence 1760; P23; P1 birth 1900, residence 1950, death `AFT 1950` \| residence 2030. Each conflicted at base and at `1885e198`; each agrees at `b13fddbb` |
| VER-5 | CON | major | An open record that agrees with a closed one (birth 1840 + birth `BEF 1850`; death 1850 + death `AFT 1841`) now removes the closed bound | To user | Probes P2, P4, P17, P18: conflict at base, agree now; the controls without the open record still conflict. P3 (death 1 Mar 1850 + burial `AFT 5 Mar 1850`) also depends on whether a burial bounds the death |

**Orchestrator's read:** three rounds of rule patches (F-18 → F-22) each fixed one case and broke another. The window needs a model, not another step. Proposal to the user:
- Unknowns: birth B and death D.
- Every event lies in [B, D].
- D − B ≤ MaxYears.
- Records for the same event are intersected when they agree and treated as alternatives (loosest) when they do not.
- The registry says per Ends role whether it fixes D (death) or only bounds it (burial: D ≤ burial).
- Propagate the bounds.
- The window is [min B, max D]; a side with no bound stays open.
- The test matrix is every counterexample so far: W1–W9, C1–C6, G1–G3, P1–P23 and the F-18 – F-20 cases.

The PR head is currently worse than base for VER-4/VER-5 pairs. The other option is to revert F-21/F-22 (back to `1885e198`'s window, keeping the user's original span fix) and do the model as its own PR.

## Round 10 — user decision (2026-10-22)
The user: lifespans are guidance to refine suggestions, always a bit fuzzy; only remove the really egregious cases (born 1950 vs died 1850); be very forgiving; find a simpler way. This answers VER-4 and VER-5 (no exact model) and replaces the F-18 – F-22 rules.

### The rule (orchestrator)
A side's lifespan is built from its **exact dates only**. A date with an open end (`BEF`, `AFT`, a half-open range) is skipped: it is too vague to rule anything out.
- **Span:** earliest first end to latest last end of the counted dates.
- **Start:** the earliest counted Starts (birth) first end, else the span's first end − `MaxYears`.
- **End:** the latest counted Ends (death, burial) last end, else the span's last end + `MaxYears`.
- **Window:** [start − `SlackYears`, end + `SlackYears`]. `SlackYears` is a new registry knob, 10 (estimate: forgiving margin for misdated records).
- **Undated:** no counted date. **Contradicts:** a counted date outside its own window, or the window inverted. With the slack, only egregious cases.
- Comparison is unchanged (`subjectcompare.lifespan`: each span against the other's window).

### Fix plan 10
**F-23 — A forgiving lifespan from exact dates, with slack** (user decision; supersedes F-18 – F-22)
Files: `core/subjectprofile/subjectprofile.go` (`lifespan`, doc comment), `core/subjectcompare/subjectcompare.go` (`Lifespan.SlackYears`), `core/matchregistry/subject.go` (`SlackYears: 10, // estimate: forgiving margin for misdated records`), `docs/matching-model.md` (knob row; update the `MaxYears` row and the lifespan description), `core/subjectprofile/entity_test.go`, `core/subjectprofile/lifespan_test.go`. Skill: `add-match-rule` (a new knob: registry, inventory row, `TestKnobInventoryMatchesRegistry`).
Do: rewrite `lifespan` to the rule above. It should be shorter than today: no collapsed/raw ends, no inverted-span fallback, no `startOpen`/`endOpen`. Delete the F-18 – F-22 test cases (W*, C*, G*, P*, the F-18 – F-20 cases) and replace them with this table in `TestEntityCompare` (lifespan outcome):

| # | Side A | Side B | Want |
| --- | --- | --- | --- |
| L1 | birth 1950 | death 1850 | conflict |
| L2 | death `BEF 11 Jan 2021` | birth 27 Jul 1932, death 31 Mar 2001 | not conflict (A undated) |
| L3 | birth 1900, death 1980 | birth 1905, death 1975 | not conflict |
| L4 | residence 1900 | residence 2030 | conflict |
| L5 | residence 1900 | residence 2015 | not conflict |
| L6 | birth 1900 | residence 1895 | not conflict |
| L7 | birth 1900 | residence 1880 | conflict |
| L8 | death 1850 | residence 1855 | not conflict |
| L9 | death 1850 | residence 1875 | conflict |
| L10 | birth `BEF 1900`, death `AFT 1950` | birth 1700 | not conflict (A undated) |
| L11 | birth 1900, birth `BEF 1895` | death 1700 | conflict |

Update the S11-32 cases in `lifespan_test.go` to the new rule. List each changed expectation with one line on why.
Check: table passes; `go test -tags fts5 ./core/...` PASS (including the knob inventory test); scripts/check-pr.sh PASS. Replace the Landed note's round 8 and 9 sentences with: "Ship rounds 8 – 10, at the user's request: a lifespan is guidance built from exact dates only, with a 10-year slack on its window, so only egregious cases conflict; open-ended dates are skipped."

## Round 10 — fix
Head `bdbfabfa`. Gate PASS (orchestrator re-ran). F-23 done: `lifespan` rewritten (open-ended dates skipped, slack window), `Lifespan.SlackYears` = 10 with its `matching-model.md` row; L1–L11 pass; the F-18 – F-22 cases are deleted. Fixtures in `claimevidence`, `promotealign` and `graphalign` moved past the slack so they still conflict (dates only). The orchestrator read the new function. No verifier: the user asked for a forgiving heuristic, not an exact model, and the table is the spec. VER-4 and VER-5 were answered by the user's decision.

## Result (after round 10)
Gate PASS · CI see PR · open: none · to user: term origin (a user term keyed like a product term reads as the product term) · manual check §10 not run

## Round 11 — user requests (2026-10-22)
1. Fix the term-origin gap on this PR. A term is unique by (property, key, origin), so a researcher's own `role` term keyed `subject` currently reads as the product term: it takes the L10n label and the registry's plain wording.
2. Remove the lifespan row from the evidence sheet. It is used only to rank suggestions, and it cannot be pinned, so it does not belong among the evidence. This supersedes the plan's AC-4.1 – AC-4.3 and the Lifespan Card. Record that in the Landed note.

### Fix plan 11
**F-24 — Only a product term reads as a product term** (from the user)
Files: `api/proto/engine.proto` (Observation `value_term_origin = 21`; the event-title message gains `type_origin`), regenerated `engine.pb.go` / `engine.pb.swift`, `core/database/observations/observations.go` and `api/ffi/handlers/{observations,source_read}.go` (select and send `property_terms.origin` beside the key), `core/database/grapheventtitles/grapheventtitles.go` (and any other `TypeKey` filler for event titles) plus its handler, `macos/App/Platform/GenealogyStore.swift` (`CatalogObservation.valueTermOrigin`, `CatalogEventTitle.typeOrigin`), `GoStore.swift`, `FakeStore.swift` (fill origin wherever it fills the key), `macos/App/Platform/ConclusionVocabulary.swift` (`SeededOrigin.provenencia`), `api/ffi/swift_vocabulary_test.go` (pin `SeededOrigin.provenencia` to `vocabkey.OriginProvenencia`), `macos/App/Platform/PropertyTermDisplay.swift`, `macos/App/Features/Catalog/Connect/EvidenceBridgeEdgeSummary.swift`, `macos/App/Features/Conclusions/EventTitleDisplay.swift`, `macos/App/Features/CitationComposer/CitationComposerVocabulary.swift`.
Failing tests first:
- Swift `PropertyTermDisplay`: a `role` term keyed `subject` with user origin and label "Main person" shows "Main person"; the product one shows the L10n label.
- `EvidenceBridgeEdgeSummaryTests` `aUserTermKeyedLikeAProductTermReadsGenerically`: a role observation with key `subject` and user origin, at a birth, reads "Alice participated as Main person at …", not "Alice was born". The same holds for a user `event_type` term keyed `birth` with a product `subject` role.
- Go observations and handler tests return `value_term_origin`; the event-titles test returns `type_origin`.
Do:
- `PropertyTermDisplay.name` takes `origin`. It consults `L10n.PropertyTerm` only when the origin is `SeededOrigin.provenencia`, so all four callers pass the origin they have.
- `EvidenceBridgeEdgeSummary`'s term key and event-type key are nil unless the term's origin is the product origin, so the registry only ever matches product keys.
- Replace the literal `"event_type"` in `EventTitleDisplay` with `SeededPropertyKey.eventType` (missed by F-16).
Check: tests pass; `grep -rn '"event_type"\|"role"' macos/App --include=*.swift` prints nothing outside `ConclusionVocabulary.swift`; scripts/check-pr.sh PASS.

**F-25 — The lifespan leaves the evidence sheet; matching keeps it** (from the user)
Files: `core/claimevidence/{claimevidence,build}.go` and tests, `api/proto/engine.proto` (drop `PromoteClaimEvidence.lifespan` and `PromoteClaimEvidenceLifespan`; `reserved 5;` and `reserved "lifespan";`), regenerated proto code, `api/ffi/handlers/promote_evidence.go` and tests, `core/database/promotealign/evidence.go` and tests, `macos/App/Platform/{GenealogyStore,GoStore,FakeStore}.swift`, `macos/App/Features/Promote/{PromoteEvidenceSheet,PromoteEvidenceDisplay}.swift`, `macos/App/Platform/L10n.swift`, `Localizable.xcstrings`, Swift tests, `docs/deployment-plan/spike-11/claim-evidence.md`, `docs/matching-model.md` (if it says the sheet shows the lifespan), the S11-20 plan's Landed note.
Do:
- The claim evidence no longer carries a lifespan, end to end: remove `claimevidence.Lifespan`, the wire message and field, the Swift types and mapping, `EvidenceLifespanCard`, `PromoteEvidenceDisplay.lifespanSentence`, and the lifespan clause of `hasConflict` and `hasNothingToCompare`.
- Remove every L10n key and xcstrings entry left without a reader: the lifespan sentences, the card heading, the "Worked out from dated events…" caption, and the spoken lifespan label.
- Remove `LifeSide.First`/`Last` and `PersonLifeDisplay.span` if nothing else reads them (grep to prove it; the Person page may still use `span`).
- Matching is untouched: `subjectprofile.lifespan`, `subjectcompare.lifespan`, the knobs and their tests stay.
- Identifier `promote.sheet.lifespan` goes.
- `claim-evidence.md`: the lifespan is a matching signal only, not a sheet line.
- Landed note: "Ship round 11, at the user's request: the lifespan is no longer on the evidence sheet (AC-4.1 – AC-4.3 withdrawn); it only ranks suggestions. A product term reads as one only when its origin is the product's."
Check: `grep -rni "lifespan" macos/App api/ffi/handlers core/claimevidence` prints only matching-side code, if any; `go test -tags fts5 ./...` PASS; scripts/check-pr.sh PASS.

## Round 11 — fix
Head `91b40b61` (F-24 and F-25 in one commit: they share the regenerated proto code and `GenealogyStore.swift`). Gate PASS (orchestrator re-ran). Beyond the file list:
- The event-header path got an origin too (`ConclusionTerm.origin = 4`, filled for Event headers only).
- `LifespanMissing` (Swift), `PersonLifeDisplay.span` and the sheet-only life plumbing in `promotealign` were removed.
- Literals in `FakeStore`, `CitationConnections` and `GenealogyStore` moved to `SeededPropertyKey`.
Failing-first was not run.

## Round 11 — review
Reviewers: MNT, CON (Sonnet), scope `1152a19e...91b40b61`. Not sent: UX (no new chrome; strings only removed), SEC+PRF (response-only fields, no new input or query shape).

| ID | Lens | Severity | Title | Decision | Evidence / reason |
| --- | --- | --- | --- | --- | --- |
| CON-1 | CON | major | Composer reading sentences, `displayLabel`, `inverseTerm`, and the Person gender treatment still treat a researcher's term as the product's | Fix now → F-27 | verified: `CitationComposerAuthoring.swift:651–672,723`; `CitationComposerAuthoringBox.swift:808–815`; `PersonGenderTreatment.swift:75–86`. None reads the origin. `ConclusionTerm` has no origin on `conclusion_detail.go:152`. Same defect as F-24 |
| MNT-1 | MNT | major | `SeededOrigin.provenencia` duplicates `CatalogOrigin.provenencia`, and the product-origin test is written out twice | Fix now → F-26 | verified: `OriginMarkers.swift:9–11` (about 34 uses); new enum `ConclusionVocabulary.swift:253`; literal at `PromoteEvidenceSheet.swift:243`. Same-pass cleanup |
| MNT-2, CON-2 | MNT, CON | major | `LifeSide` / `FirstLifespanLead` / the second return of `Build`/`ForSubject` are dead | Fix now → F-28 | verified: every non-test caller discards it with `_`; F-25 asked for the grep. Matching behavior unchanged |
| MNT-3, CON-3 | MNT, CON | major | `add-match-rule` SKILL.md, `comparison-layers.md:385` and the plan's AC table still describe the sheet's lifespan | Fix now → F-29 | verified at the lines quoted |
| MNT-5 | MNT | minor | `ConclusionTerm.origin` filled only for Event headers | Fix now → F-27 | the gender fix needs it everywhere |
| MNT-4 | MNT | minor | Second full scan of `property_terms` in `conclusionheaders/compose.go` | Fix now → F-27 | one scan of id, label, origin |
| CON-4 | CON | minor | `searchindex/handles.go:371` uses `Parts.TypeKey` as an English fallback search word | Rejected | a search word, not a claim that the term is the product's |
| CON-5 | CON | minor | F-24's file list named a file that needed no change | Rejected | no defect |

### Fix plan 12
**F-26 — One product-origin constant and one test** (from MNT-1)
Files: move `CatalogOrigin` from `macos/App/Features/Catalog/OriginMarkers.swift` to `macos/App/Platform/ConclusionVocabulary.swift`; delete `SeededOrigin`; add `CatalogOrigin.isProduct(_ origin: String) -> Bool`. `api/ffi/swift_vocabulary_test.go` pins `CatalogOrigin.provenencia` and `.user` to Go. `PropertyTermDisplay`, `EvidenceBridgeEdgeSummary` and every reader in F-27 call `isProduct`. Replace the literal at `PromoteEvidenceSheet.swift:243`.
Check: `grep -rn 'SeededOrigin\|"provenencia"' macos/App` prints nothing outside `ConclusionVocabulary.swift`. `grep -rn 'origin == CatalogOrigin.provenencia\|== CatalogOrigin.provenencia' macos/App` prints only the predicate's body, or uses where a whole-row origin (not a term's) is shown, such as the origin pills: list those and leave them.

**F-27 — Every term-key reader honours the origin** (from CON-1, MNT-4, MNT-5)
Files: `macos/App/Features/CitationComposer/CitationComposerAuthoring.swift`, `CitationComposerAuthoringBox.swift`, `macos/App/Features/Conclusions/PersonGenderTreatment.swift`, `api/ffi/handlers/conclusion_detail.go` (fill `ConclusionTerm.origin` wherever a `ConclusionTerm` is built; drop the "Event header only" comment), the Swift `.term` case of `CatalogConclusionValue` and its mapping, `core/database/conclusionheaders/compose.go` (one scan of id, label, origin), tests.
Failing tests first:
- `inverseTerm` of a product `parent` never returns a researcher's `child`.
- A researcher's `relationship_type` keyed `spouse` does not get the "are spouses" sentence.
- A researcher's `place_relationship_type` keyed `part_of` keeps its own label.
- A researcher's `gender` term keyed `woman` gets the `.other` family.
- `conclusion_detail` returns `origin` on a term value.
Do: gate each read on `CatalogOrigin.isProduct(term.origin)`. That covers `inverseTerm`'s source and its lookup.
Check: `rg -n 'term\.key ==|switch term\.key|\.key == Seeded|inverseTermKey\[' macos/App` prints only reads gated by `isProduct`; scripts/check-pr.sh PASS.

**F-28 — Drop the dead lifespan side** (from MNT-2, CON-2)
Files: `core/subjectprofile/subjectprofile.go`, `core/subjectcompare/subjectcompare.go`, `core/database/promotealign/{layer,profiles,evidence}.go`, `core/graphalign` and `core/subjectprofile` tests.
Do: `lifespan` returns `(window, span propertycompare.Value, ok bool)`. `Build`, `ForSubject` and `addLifespans` lose the second return. Delete `LifeSide`, `FirstLifespanLead` and the exported `LifespanMissing` if nothing else reads them (grep). Update callers and the side assertions in `lifespan_test.go` to assert `ok`, window and span. Matching behavior is unchanged: all matching tests pass unedited, apart from the signature.
Check: `grep -rn 'LifeSide\|FirstLifespanLead' core api` prints nothing; `go test -tags fts5 ./...` PASS.

**F-29 — Docs say the lifespan is a matching signal only** (from MNT-3, CON-3)
Files: `.agents/skills/add-match-rule/SKILL.md` (lines 34, 75), `docs/deployment-plan/spike-11/comparison-layers.md:385`, the S11-20 plan: mark AC-4.1 – AC-4.3, D-13, STEP-8, T-3, the `promote.sheet.lifespan` identifier and the lifespan strings rows "(withdrawn, ship round 11)".
Check: `grep -rn 'Evidence.Lifespan\|LifespanMissing\|LifeSide' .agents docs` prints only withdrawn-marked plan rows; `scripts/check-plan-docs.py` passes.

## Round 12 — fix
Head `b4541135` (`c751c621` F-26, `284ed414` F-27, `85daf9d4` F-28, `b4541135` F-29). Gate PASS (orchestrator re-ran). The F-26 vocabulary pin and the F-27 Go and Swift tests failed first; F-28 is a refactor. The landed S11-32 plan is left as the record of what it shipped.

## Round 12 — review
MNT and CON (Sonnet), scope `91b40b61...b4541135`. CON-1, CON-2, MNT-1 – MNT-5 all **Resolved**; no new blockers. Product terms keep their wording, inverses and gender families. Matching's lifespan output is unchanged (only the dead return was dropped).
Nits (logged): `MNT-6 · conclusionheaders/compose.go:72-104 · scanTerms repeats scanVocab's loop` · `MNT-7 · ConclusionVocabulary.swift:253-258 · CatalogOrigin doc names only source tables`

## Finish
Merged `origin/main` @ `d36060a0` (S11-26 plan). One conflict in the map (`deployment-plan.md`): kept S11-20 "Landed #525" and S11-26's new plan link and "Planned". Gate PASS on `e4a63fb5`.

## Result (after round 12)
Gate PASS · CI see PR · open: none · to user: — · manual check §10 not run · follow-ups: none

## Round 13 — user report (2026-10-22)
Two Swift 6 concurrency warnings from the user's build: `ConnectionSentenceRegistry.entries` (line 45) and `.sentences` (line 90) are static arrays of non-`Sendable` structs (their closure fields). A clean `build-for-testing` (scratch derived data) shows these as the only warnings in Swift files this PR changed; the other 101 are in files it does not touch. The gate missed them because `xcode-test` builds incrementally and prints no warnings for files it does not recompile.

**F-30 — The registry's rows are Sendable** (from the user)
Files: `macos/App/Features/Catalog/Connect/ConnectionSentenceRegistry.swift`.
Do: `Slots`, `KindEntry` and `Sentence` conform to `Sendable`; every closure field is `@Sendable` (`(Slots) -> String` → `@Sendable (Slots) -> String`, likewise `phrase` and `noTerm`). No `nonisolated(unsafe)`, no `@MainActor` on the enum. Nothing else changes.
Check: a clean build with `-derivedDataPath` in the scratch dir prints no warning for `ConnectionSentenceRegistry.swift`; scripts/check-pr.sh PASS.
F-30 done at `0e5a543d`: `Slots`, `KindEntry`, `Sentence` are `Sendable`, closure fields `@Sendable` (and the private `subjectSentence` helper's parameter). A `build-for-testing` that recompiled the file prints no registry warning; gate PASS (orchestrator re-ran).

## Result (after round 13)
Gate PASS · CI see PR · open: none · manual check §10 not run

## Round 14 — user requests (2026-10-22)
1. The title badges' tooltip must match the design. The board (`Promote flow.dc.html`, `titleNode`) wraps the marks in `PVTooltip` (label = the row's reason, side left; marks focusable, help cursor). S11-D2 §5 gives the reason as "via John Smith → PER-7QH3D (spouse)", "(subject)", "(part of)": the role is the connection's term, lowercase.
   - **Today:** `.help(reason)` is the kit's macOS tooltip (README Tooltip row), so presentation is right. The **text** is wrong: `PromoteModel.reasonText`'s `.via` puts `BridgeFact.phrase` in the parentheses. That phrase is the graph card phrase from the registry ("Participated as Subject", "Is the spouse of"). This predates S11-20; the Promote row's reason uses the same text, so both change.
2. A click outside the evidence sheet closes it. Today it is a plain `.sheet` (`PromoteView.swift:68`). The outside-click watcher exists only privately in `PVFormDialog` (`PVFormDialogOutsideClick`).

### Fix plan 14
**F-31 — The reason names the connection by its term** (from the user)
Files: `macos/App/Features/Catalog/Connect/ConnectionSentenceRegistry.swift`, `EvidenceBridgeEdgeSummary.swift`, `macos/App/Features/Promote/{PromoteFlow,PromoteModel}.swift`, tests.
Failing test first: `PromoteModelTests` `aViaReasonNamesTheTermInLowercase`. A participation bridge with role Subject gives "via John Smith → PER-7QH3D (subject)". A spouse relationship gives "(spouse)", a place relationship part_of gives "(part of)". A location bridge has no term, so it reads "via Leeds → PLC-6M3FD" (`viaNeighbor`, no parentheses). A researcher's term keeps its own label, lowercased.
Do:
- `KindEntry` gains `reasonRole: @Sendable (String?) -> String?`. The three term kinds return the term label lowercased with the current locale (`lowercased(with: .autoupdatingCurrent)`); location returns nil.
- `EvidenceBridgeEdgeSummary` exposes `reasonRole(for:)`, with the same origin-aware term label it already reads.
- `PromoteFlow.BridgeFact` carries `reasonRole: String?`, set in `PromoteModel.facts()`.
- `reasonText` uses it instead of `phrase`: nil or empty gives `viaNeighbor`.
- If nothing else reads `BridgeFact.phrase`, remove it (grep).
Check: test passes; scripts/check-pr.sh PASS.

**F-32 — A click outside the evidence sheet closes it** (from the user)
Files: `macos/App/DesignSystem/Components/FormDialog/PVFormDialog.swift`, a new kit file for the shared modifier (e.g. `macos/App/DesignSystem/Components/Panel/PVOutsideClick.swift`), `macos/App/DesignSystem/README.md`, `macos/App/Features/Promote/PromoteView.swift` (or `PromoteEvidenceSheet.swift`).
Do:
- Move `PVFormDialogOutsideClick` and its view out of `PVFormDialog` into one kit modifier, `.pvOutsideClick(isEnabled:perform:)`. Its doc: a click in the sheet's window outside its content runs `perform`; menus and popovers are not this path.
- `PVFormDialog` uses it, with its behavior unchanged.
- The evidence sheet's content uses it to clear the sheet binding (close; nothing to discard: pins and the claim fields already live in the model).
- Add a README line.
- No second copy of the watcher. `grep -rn "addLocalMonitorForEvents" macos/App/DesignSystem` shows one site for this purpose.
Check: scripts/check-pr.sh PASS. A unit test where the existing `PVFormDialogDismiss` tests make that possible; otherwise say so (no UI target).

## Round 14 — fix
Head `d591470d` (`7e48d6b6` F-31, `d591470d` F-32). Gate PASS (orchestrator re-ran).
- **F-31:** `aViaReasonNamesTheTermInLowercase` failed first on all five cases (subject, spouse, part of, location without a role, a researcher's "Godparent") and now passes. `BridgeFact.phrase` stays: the will-file list reads it (`PromoteFlow.swift:282`).
- **F-32:** the watcher moved to `DesignSystem/Components/Panel/PVOutsideClick.swift` as `.pvOutsideClick(isEnabled:perform:)`. `PVFormDialog` uses it unchanged. The evidence sheet calls `model.closeSheet()`. The README line is added. The helper enum was renamed `PVFormDialogClick` → `PVOutsideClick`, and the existing click tests were updated. No UI test exists for the sheet's own outside click.
- **Clean `build-for-testing`:** no warning in this PR's app files. The 12 warnings in `PVPanelFormDialogTests.swift` are on lines the PR does not touch, from before it.

## Result (after round 14)
Gate PASS · CI see PR · open: none · manual check §10 not run

## Round 15 — user request (2026-10-23)
The connection sentences print the term label capitalized mid-sentence ("John Smith is the Spouse of Mary Smith", "Leeds is Part of Yorkshire", "participated as Witness at"). The user asked: lowercase the term inside the sentences, the same way the reason does. The "as built" brief posted to the design board (chat "S11-20 as built", 2026-10-23) already describes the lowercase wording.

**F-33 — A term reads lowercase inside every connection phrase** (from the user)
Files: `macos/App/Features/Catalog/Connect/ConnectionSentenceRegistry.swift`, `EvidenceBridgeEdgeSummary.swift`, `macos/ProvenenciaTests/ConnectionSentenceRegistryTests.swift`, `EvidenceBridgeEdgeSummaryTests.swift`, and any other test that pins a capitalized term inside a sentence.
Failing tests first:
- "John Smith is the spouse of Mary Smith".
- "Leeds is part of Yorkshire".
- "Mary Smith participated as witness at Baptism of John Smith".
- The card phrases lowercase the term too, with the phrase's first word still capitalized: "Is the spouse of", "Participated as witness".
Do:
- Lowercase the term label once, where it enters the registry. Use `lowercased(with: .autoupdatingCurrent)`, as `reasonRole` does.
- Every sentence, phrase and the reason read that one value; `reasonRole` drops its own lowercasing.
- Endpoint names (people, places, event titles) are never lowercased.
Check: tests pass; `grep -n "lowercased" macos/App/Features/Catalog/Connect` shows one site; scripts/check-pr.sh PASS.
F-33 done at `6763b144`; gate PASS (orchestrator re-ran). `aTermReadsLowercaseInsideEverySentenceAndPhrase` failed first and passes. One lowercasing site (`ConnectionSentenceRegistry.termInLowercase`); `KindEntry.reasonRole` closure became `namesTerm: Bool`. Four assertions pinning capitalized terms were updated to lowercase. No new warning in the Connect files.

## Result (after round 15)
Gate PASS · CI see PR · open: none · manual check §10 not run · design board: "S11-20 as built" brief posted to the project's chat panel, not yet applied

## Round 16 — board revision (2026-10-23)
The user revised the S11-D2 board after using the built sheet with real data, and asked for a top-to-bottom re-review, with any new data requirements implemented on this PR. Board read via the Claude Design connector: `Promote flow.dc.html` (etag `1792762007305273`, the `sheets` frames and `build()`) and `EvidenceCard.dc.html` (same etag). The board chat's tail could not be read (the transcript exceeds the connector's cap), so the board files are the spec. Where the board still draws something the user later replaced, this PR's later decisions stand: no lifespan line (round 11); "Show N records" wording.

### What the board now draws (diff against the built sheet)
**Title row**
- `PVPanel` title reads "<row name> *joins* <record title> <REF>" ("joins" muted italic; the ref in mono, muted).
- The band and Contested badges follow the ref inline, with the reason as their tooltip.
- A ghost close button (`x`, "Close") sits at the far right of the title row.
- The footer keeps "Back to all rows".

**Subtitle (new):** "This is an identity claim to join **<row name>** from the *<Source title>* into the concluded <person|event|place> **<record title>**. Check the box on each observation below that supports the claim. Observations that clearly match are already checked — review them before you file."
- Names: semibold, display ink. The Source title: italic, link ink. Same weights in the evidence heading.

**Under the title:** the Updated and contested lines, then the claim fields. Unchanged.

**Evidence heading**
- A divider precedes the evidence.
- The section header reads "**<row name>** from the *<Source title>* belongs in the concluded <kind> **<record title>** because…", with meta "N checked".
- The intro sentence ("Check every line…") is **gone**. The conflict Callout ("Some evidence disagrees") is **gone**.

**Property card header (sunken):** "…they have a similarity in <property label, lowercase>" (Newsreader 14/500), with "This source: <value>" after it.

**Line grid:** box 18 · value 200 · sentence (flex) · source 200 · badge 116.
- **Value:** the first member's value (primary, 500). A likeness line shows its two 96 pt crops here, gap 8, no captions.
- **Sentence:** secondary ink, 400. Below it, "This source: X" only when the line's incoming value differs from the card's.
- **Source:**
  - one member: the record's Source title, italic and muted;
  - several members: a ghost disclosure, "Show N records" / "Hide N records".
- The "· N records" suffix on the sentence is **gone**.

**Expanded members:** grid box · value 170 · (empty) · source 200, italic and muted · (empty), indented 44. **No box is ever disabled.**

**Relationship card header (sunken):** box · "…they share <phrase> (<this-source ref> and <record ref>)" with " · N records" when bundled · outcome badge.
- The phrases by board example:
  - "a birth event" (the neighbor is an event; its type, lowercase);
  - "a subject, John Smith" (the neighbor is a person; the role, then the name);
  - "a spouse";
  - "a parent place".
- A similar outcome inserts "similar": "a similar marriage event".
- With no named connection, "a connection".
- The refs: the neighbor row's ref on this Source (CEV-3B5RT) and its record's handle ref (EVT-4RJ9T).
- Group labels under the relationship: "…they have a similarity in <property>" (Newsreader 13, secondary) + "This source: X", indented 28.
- "Nothing else to compare." stays.

**Pinning (the data requirement):** board notes E03 and E05 say a conflicting line starts unchecked and *can still be pinned*. `build()` never disables a box. Today Go marks only checked, `evidence.Pinnable` partial, and by-eye lines pinnable. So:
- the pin policy changes in Go: every line, member and relationship can be pinned;
- the default check is unchanged (`evidence.Checked`, by eye).

**States:** loading, failed, New, Skip and nothing-to-compare are unchanged, apart from the heading copy.

### Fix plan 16
**F-34 — Every line can be pinned; Go owns it** (board E03/E05)
Files:
- Go: `core/claimevidence/{claimevidence,sections,build}.go`, `core/database/promotealign/pins.go` (+ tests), `core/matchregistry/evidence.go` (+ test).
- Docs and wire: `docs/matching-model.md` (drop the `evidence.Pinnable` row), `docs/deployment-plan/spike-11/claim-evidence.md` (Lines: every line pinnable, conflicts start unchecked), `api/proto/engine.proto` (drop `pinnable` on the line and relationship messages; reserve their numbers and names).
- Swift: regenerated code, `api/ffi/handlers/promote_evidence.go`, `GenealogyStore`/`GoStore`/`FakeStore`, and Swift tests.
Failing tests first:
- Go: `PinOffers` offers a conflicting line's member ids.
- Go: `ResolvePins` keeps a checked conflict id.
- Go: the default pins still leave a conflict unchecked.
Do:
- Delete `Line.Pinnable`, `Relationship.Pinnable` and the `Rules.Pinnable` knob, and every reader. `PinOffers` offers every line's ids.
- In Swift, delete `memberDisabled`, the disabled-box logic and `hasConflict`.
- Knob inventory and matching tests updated.

**F-35 — The sheet as the board draws it** (board E01 – E14)
Files: `macos/App/Features/Promote/{PromoteEvidenceSheet,PromoteEvidenceDisplay}.swift`, `macos/App/DesignSystem/Components/Panel/PVPanel.swift` (only if the title row needs it), `macos/App/Platform/L10n.swift`, `Localizable.xcstrings`, Swift tests. Skills: `add-ui-component`, `add-localized-string`.
Do, in the order of the section above:
1. **Title row:**
   - "<name> joins <record title> <REF>", then the badges (tooltip and spoken label unchanged), then the close button (`PVIconButton`, `x`, label "Close", calls `model.closeSheet()`, identifier `promote.sheet.dismiss`).
   - If `PVPanel` cannot place inline badges after the title text, give it a title-view initializer. One kit change, documented in the README.
2. **Subtitle and heading:**
   - Each is one localizable format with styled arguments: names semibold in display ink, the Source title italic in link ink. One helper builds the `AttributedString` from a format and styled arguments, so translators can reorder.
   - Kind word: person / event / place from the row's kind.
   - The Source title is `model.sourceTitle`. When it is nil, use "this source" (a new key) in plain ink.
   - Delete the intro sentence and the conflict Callout with their keys.
3. **Card headers:**
   - "…they have a similarity in %@": the property's display label, lowercased with the current locale.
   - Show "This source: X" in the header when every line in the card has the same incoming value; otherwise per line, under the sentence.
4. **Line grid as specified:**
   - The value is the first member's worded value; a likeness shows its crops.
   - Source column: the single member's Source title, or the "Show N records" / "Hide N records" disclosure (new "Hide %lld records" plural key).
   - Remove the count suffix.
   - Expanded member rows use the board's columns.
   - Spoken labels keep their content: card title, sentence, count when bundled, this source.
5. **Relationship header: a registry-driven share phrase.**
   - Add to `ConnectionSentenceRegistry` a `share` wording per kind and the neighbor's side, with an agreeing and a similar format each (so "similar" is never spliced in):
     - participation, the neighbor is the event: "a %@ event" / "a similar %@ event" (its event-type label, lowercase); with no type, "an event" / "a similar event";
     - participation, the neighbor is the person: "a %1$@, %2$@" (role lowercase, name);
     - relationship: "a %@" with the neighbor's role relative to the row. That is the term when the neighbor is the A end; otherwise its inverse. Move `CitationComposerAuthoring.inverseTermKey` into `ConclusionVocabulary.swift` as the one inverse map (same-pass), with spouse, sibling and cousin as their own inverse;
     - place relationship: part_of with the neighbor as the containing place gives "a parent place"; the other end gives "a place within it"; succeeded_by gives "a successor place" / "a predecessor place";
     - location: "a place, %@".
     - Fallback: "a connection" / "a similar connection".
   - The header reads "…they share %1$@ (%2$@ and %3$@)": phrase, the neighbor row's own ref, then the record ref. With no neighbor row, only the record ref: "(%@)".
   - This replaces `relationshipHeader` and its fallback lines on the sheet. The bridge sentence stays for the graph and the page.
6. Group labels under relationships use the same "…they have a similarity in %@".
7. Delete keys and helpers left without a reader (grep).
Failing tests first, in `PromoteEvidenceDisplayTests` and `ConnectionSentenceRegistryTests`:
- the subtitle and heading formats with and without a Source title;
- `…they have a similarity in name`;
- "This source" placement: one value → header; mixed → lines;
- each share phrase above, including similar, inverse and fallback;
- the relationship header with two refs and with one;
- "Hide 4 records".
Check: tests pass; `grep -rn "conflictCallout\|evidenceIntro\|memberDisabled\|hasConflict" macos/App` prints nothing; scripts/check-pr.sh PASS; a clean build shows no new warning in changed files.

## Round 16 — fix
Head `8aab38bf` (`0818810b` F-34, `8aab38bf` F-35). Orchestrator gate: the first run FAILed in `EvidenceCardLayoutTests/measuredLayoutMatchesThePaint("uncited")`, a 1.69 pt height mismatch in a graph card this round does not touch. The next two runs PASSed. Logged as flaky.
- **F-34:** the Go tests failed first.
- **F-35:** tests written after the code (the implementer's deviation 1).
Deviations accepted:
- untitled-Source formats instead of a "this source" argument;
- "joins" its own key;
- location share phrase only from the place side;
- one ref when the neighbor has no record or the refs match;
- `PVPanel.titleAccessoryPlacement: .afterTitle`;
- `Platform/StyledFormat.swift`;
- `BridgeFact.termKey`/`eventType`.
No new warning in changed files (the implementer's clean build).

## Round 16 — review
Reviewers: CON (Sonnet), MNT, UX. Scope `0b2f4dab...8aab38bf`. Skipped SEC+PRF: `promotealign/pins.go` only drops a filter. There is no new input, loop or query.

| ID | Lens | Severity | Title | Decision | Evidence / reason |
| --- | --- | --- | --- | --- | --- |
| UX-1 | UX | major | Share phrases put "a" before vowel-initial terms ("a informant", "a engagement event", "a aunt / uncle") | Fix now → F-36 | verified: `L10n.swift:2270,2284,2291` fixed "a %@…"; seeded `informant`, `engagement`, `pibling` |
| UX-2 | UX | major | Source titles cut off in the 200 pt column have no tooltip (likeness regression of `LikenessTile`'s `.help`) | Fix now → F-37 | verified `EvidenceSourceTitle` `lineLimit(2)`, no `.help` |
| CON-1, MNT-4 | CON, MNT | major | Skill and four docs still describe `evidence.Pinnable` / "not pinnable" | Fix now → F-38 | verified by grep at the lines quoted |
| MNT-1 | MNT | major | `PVPanel` has two accessory placements; `.trailing` has no caller | Fix now → F-37 | verified: only caller passes `.afterTitle` |
| MNT-2 | MNT | major | `StyledFormat` / `AttributedString: PVCopy` not in `add-localized-string` or README; tests not beside the helper | Fix now → F-38 | new pattern, Q10 |
| MNT-3 | MNT | major | Swift inverse map mirrors Go's seeded `Inverse` with no parity test | Fix now → F-38 (parity test); Follow-up: carry `inverse_key` on the wire term (`docs/ideas/term-management.md` item 2) | `.cursor/rules/macos-client.mdc:20` |
| MNT-5 | MNT | major | `eventTypeLabel` re-implements `eventTypeKey`'s lookup, and they pick endpoints differently | Fix now → F-36 | verified `EvidenceBridgeEdgeSummary.swift:31-39` vs `:122-130` |
| UX-3, CON-4 | UX, CON | minor | Heading names not set apart by weight (section header is all semibold) | Fix now → F-37 | spec: names semibold against lighter text |
| UX-4 | UX | minor | Long title may wrap at half width beside a flexible accessory | Fix now → F-37 | layout priority |
| UX-5 | UX | minor | Untitled Source: blank on a line, "a record" on a likeness | Fix now → F-37 | one rule |
| UX-7 | UX | minor | Spoken relationship label starts with "…" | Fix now → F-36 | spoken variant without the ellipsis |
| MNT-6 | MNT | minor | `PromoteEvidenceDisplay` owns fonts and colors | Fix now → F-37 | file this PR adds |
| MNT-7 | MNT | minor | Group label re-draws `EvidenceCardHeader` | Fix now → F-37 | |
| MNT-8 | MNT | minor | `BridgeFact.reasonRole` read for the share phrase | Fix now → F-36 | rename to `term` |
| MNT-9 | MNT | minor | `endTitles` switches on `"part_of"`/`"succeeded_by"` literals | Fix now → F-38 | file this PR changed; seeded keys belong in `ConclusionVocabulary` |
| PRF (MNT outside) | PRF | minor | A linear subject scan per bridge on each page build | Fix now → F-36 | one id→subject dictionary |
| UX-6 | UX | minor | "from the The Times…" when a Source title starts with "The" | To user | board copy |
| CON-2 | CON | minor | Two same-kind connections between one pair both take the first one's phrase | Follow-up | not new; needs the relationship to carry its bridge id |
| CON-3 | CON | minor | A researcher term with no inverse falls back to "a connection" on the B side | Rejected | accepted deviation; F-36 gives researcher terms an article-free form |
| CON-5 | CON | minor | F-35 tests after code | Logged | accepted deviation |

Nits (logged): none beyond the minors above.

### Fix plan 17
**F-36 — Share phrases carry their own article; one event-type lookup** (from UX-1, MNT-5, MNT-8, UX-7, PRF)
Files: `ConnectionSentenceRegistry.swift`, `EvidenceBridgeEdgeSummary.swift`, `PromoteFlow.swift`, `PromoteModel.swift`, `PromoteEvidenceSheet.swift`, `L10n.swift`, `Localizable.xcstrings`, tests.
Failing tests first:
- "…they share an informant, John Smith";
- "…they share an engagement event";
- "…they share an aunt or uncle" (pibling, neighbor at the A end; the seeded label "Aunt / uncle" reads "aunt or uncle" in the phrase key);
- "…they share a similar engagement event";
- a researcher role, event type or relationship term reads "a connection (Godparent)" / "a similar connection (Godparent)";
- the spoken relationship label has no leading ellipsis.
Do:
- Each seeded role, event type and relationship term gets its own share phrase key, with the article in the key, in the registry's term table. Use `Seeded…Term` constants, never literals.
- The similar form is "a similar %@" over a per-term bare noun key, so English needs no article logic.
- A term with no seeded phrase (a researcher's term) uses the article-free fallback above.
- The shared share formats that spliced "a %@" are deleted.
- Extract `eventTypeObservation(subjectID:snapshot:)`. `eventTypeKey` and `eventTypeLabel` both build on it and choose the endpoint the same way (A, then B).
- Build one id→placed-subject dictionary per snapshot and use it in the summary's lookups during `facts()`.
- Rename `BridgeFact.reasonRole` to `term`, documented for both readers.
- The spoken label drops the leading "…".

**F-37 — Sheet polish and one PVPanel placement** (from UX-2, MNT-1, UX-3, UX-4, UX-5, MNT-6, MNT-7)
Files: `PVPanel.swift`, `macos/App/DesignSystem/README.md`, `PromoteEvidenceSheet.swift`, `PromoteEvidenceDisplay.swift`, tests.
Do:
- `EvidenceSourceTitle` gets `.help(Text(verbatim: title))`.
- An untitled Source shows "a record" (`untitledRecord`) on every line and member row, likeness or not.
- `PVPanel` draws the accessory after the title, always. Delete `PVPanelTitleAccessoryPlacement` and its parameter. The title gets `.layoutPriority(1)`. Update the README row: "after the title, filling the row; put a Spacer in it to push a control right".
- The heading's names and record title are semibold display ink over a regular-weight base. If `PVSectionHeader` forces semibold, give it a title-weight option or let an `AttributedString` title carry its own weights, one kit change, documented.
- `PromoteEvidenceDisplay` returns words and semantic roles (name, source, muted, ref). The sheet maps the roles to fonts and colors. No SwiftUI import in the display helper.
- `EvidenceCardHeader` gains a style (card / group). The relationship group label uses it.
Check: scripts/check-pr.sh PASS; `grep -n "import SwiftUI" macos/App/Features/Promote/PromoteEvidenceDisplay.swift` prints nothing; `grep -rn "titleAccessoryPlacement" macos/App` prints nothing.

**F-38 — Docs, skills and vocabulary guards** (from CON-1, MNT-4, MNT-2, MNT-3, MNT-9)
Files: `.agents/skills/add-match-rule/SKILL.md`, `docs/deployment-plan/spike-11/{claim-evidence,comparison-layers,deployment-plan,completed}.md`, `.agents/skills/add-localized-string/SKILL.md`, `macos/App/DesignSystem/README.md`, `macos/ProvenenciaTests/StyledFormatTests.swift` *(new, moved from `PromoteEvidenceDisplayTests`)*, `api/ffi/swift_vocabulary_test.go`, `macos/App/Features/CitationComposer/CitationComposerAuthoring.swift`, `docs/ideas/term-management.md`.
Do:
- Remove `evidence.Pinnable` from the skill's inventory list. Reword the four doc lines to "every line can be pinned; only the default check depends on the relation".
- Add a "Styled arguments" section to `add-localized-string`: when to use `StyledFormat.attributed`, its `%@` / `%n$@` / `%%`-only grammar, and that `AttributedString` is accepted wherever `PVCopy` is. Mention `AttributedString` in the README's `PVCopy` line.
- Add a Go subtest that reads `SeededRelationshipTypeTerm.inverse` and compares it with `subjectvocab`'s `Inverse` (undirected terms map to themselves). Add one line to `docs/ideas/term-management.md` item 2: carry `inverse_key` on the wire term and delete the Swift map.
- `endTitles` switches on `SeededPlaceRelationshipTypeTerm` constants.
Check: `git grep -n -i "evidence.Pinnable\|not pinnable" -- ':!*.ship.md' ':!docs/deployment-plan/spike-11/design/archive' ':!docs/deployment-plan/spike-11/prs/S11-3*' ':!docs/deployment-plan/spike-11/prs/S11-19*' ':!docs/deployment-plan/spike-11/prs/S11-25*'` prints nothing; `go test -tags fts5 ./api/ffi/...` PASS; scripts/check-pr.sh PASS.

## Round 17 — fix
Head `a5007da0` (`2bda0162` F-36, `ee727c98` F-37, `a5007da0` F-38). Gate PASS (orchestrator re-ran). Failing-first was run for F-36 and F-37 (behind stubs). The F-38 inverse-parity test fails when the map is corrupted.
Deviations:
- F-38's grep also cleared the landed S11-01 plan's `evidence.Pinnable` inventory row. Accepted: it is the knob inventory others copy.
- Seeded terms outside the share tables fall back to "a connection (Label)".
- Two `"part_of"`/`"succeeded_by"` literals remain at `CitationComposerAuthoring.swift:657,660` → F-40.
- **New warnings.** A clean build (scratch derived data `dd2`) shows 244 Swift 6 warnings at `<unknown>:0`: "cannot form key path that captures non-Sendable type `KeyPath<…FontAttribute / ForegroundColorAttribute / InlinePresentationIntentAttribute>`". The clean build after round 14 had none. The attribute writes added in rounds 16–17 (`StyledFormat.swift`, `PromoteEvidenceDisplay.swift`, `PromoteEvidenceSheet.swift`, `StyledFormatTests.swift`) run outside the main actor. The pre-existing `PVComboBox` writes run in a `View` and do not warn. → F-39.

### Fix plan 18
**F-39 — Styled text builds without Swift 6 warnings** (orchestrator, from the clean build)
Files: `macos/App/Platform/StyledFormat.swift`, `macos/App/Features/Promote/{PromoteEvidenceSheet,PromoteEvidenceDisplay}.swift`, `macos/ProvenenciaTests/StyledFormatTests.swift`.
Do:
- Make every place that sets an `AttributedString` attribute through a key path run in a main-actor context, as `PVComboBox` does: mark the style-mapping functions `@MainActor` (or the type that owns them), and the tests that call them.
- Alternatively, apply attributes without dynamic key paths, if that is simpler and keeps one way.
- No `nonisolated(unsafe)`, no `@preconcurrency` suppression.
- `PromoteEvidenceDisplay` stays SwiftUI-free. Its semantic `Role` attribute must not warn either; if it does, make it a custom `AttributedStringKey` that is `Sendable`.
Check: a clean `build-for-testing` with a fresh `-derivedDataPath` prints 0 `<unknown>:0` key-path warnings and no warning in any file this PR changed; scripts/check-pr.sh PASS.

**F-40 — No seeded-key literals left in the composer** (from MNT-9, deviation above)
Files: `macos/App/Features/CitationComposer/CitationComposerAuthoring.swift`.
Do: replace the `"part_of"` / `"succeeded_by"` literals at lines 657 and 660 (and any other seeded key literal in the file) with `SeededPlaceRelationshipTypeTerm` / `SeededRelationshipTypeTerm` constants.
Check: `grep -nE '"(part_of|succeeded_by|spouse|sibling|parent|child)"' macos/App/Features/CitationComposer` prints nothing; scripts/check-pr.sh PASS.

**F-41 — "from <Source title>", without "the"** (UX-6, user decision 2026-10-23: "Just drop the 'the' from that copy.")
Files: `macos/App/Platform/L10n.swift`, `Localizable.xcstrings`, the tests that pin the subtitle and heading.
Do:
- In the titled subtitle and evidence heading formats, "from the %2$@" becomes "from %2$@": "This is an identity claim to join **John Smith** from *Leeds Mercury obituary* into…", "**John Smith** from *Leeds Mercury obituary* belongs in…".
- The untitled variants keep "from this source".
Check: tests updated and pass; scripts/check-pr.sh PASS.

## Round 18 — fix
Head `9ad6a7f0` (`ce65f4d1` F-39, `86256b23` F-40, `9ad6a7f0` F-41). Gate PASS (orchestrator re-ran).
- **F-39:** `<unknown>:0` key-path warnings went from 244 to 0 on a fresh derived-data build. `@MainActor` did not clear them; typed `AttributeContainer` subscripts did, the way `PVComboBox` writes them.
- **F-40:** no seeded-key literal left in `Features/CitationComposer`.
- **F-41:** "from %2$@" in the titled subtitle and heading.
- **CON-1:** **Resolved**. The orchestrator's `git grep` for `evidence.Pinnable` / "not pinnable" prints nothing.

## Round 17/18 — review 2
Reviewers: MNT, UX (each had Fix-now majors). Scope `8aab38bf...9ad6a7f0`. CON's only major (docs) is checked above.
- **UX:** UX-1 and UX-2 **Resolved**; UX-3 – UX-7 done. All 59 `evidenceGraph.share.*` strings read correctly. Nits: `UX-9 · L10n.swift:1543 · untitledRecord comment says likeness` · `UX-10 · PromoteEvidenceDisplay.swift:210 · spokenHeader strips a hard-coded "…"` · `UX-11 · PromoteEvidenceDisplay.swift:69 · untitled non-likeness member spoken as the value only`.
- **MNT:** MNT-1, -2, -3, -5 **Resolved**; MNT-6 – MNT-9 done. One mechanism each for share phrases (registry tables) and attribute styling (typed `AttributeContainer` subscripts). Nits: `MNT-10 · EvidenceBridgeEdgeSummary.swift:17 · reasonRole(for:) feeds BridgeFact.term` · `MNT-11 · add-localized-string · say to use typed AttributeContainer subscripts` · `spokenHeader edits localized output` · `the sheet repeats PVSectionHeader's h3 size for the heading names` · `a researcher's event type keeps its case where roles are lowercased`.
- **Follow-up (MNT, logged):** seeded terms are listed by hand in four Swift places: the `Seeded…Term` enum, `L10n.PropertyTerm`, the share tables and the share test. The Go test checks Swift ⊆ seed, not seed ⊆ Swift, and `add-seeded-vocabulary` names none of these steps. An idea, not this PR.

## Finish (round 18)
`main` unchanged since `d36060a0`. Gate PASS on `9ad6a7f0`.

## Result (after round 18)
Gate PASS · CI see PR · open: none · to user: — · follow-ups: hand-kept seeded-term lists (idea); the relationship card takes the first same-kind bridge (CON-2); carry `inverse_key` on the wire (term-management idea) · flaky: `EvidenceCardLayoutTests/measuredLayoutMatchesThePaint("uncited")` · manual check §10 not run · design board: revised by the user; the sheet now follows it

## Round 19 — user requests (2026-10-23)
1. In the sheet's title, subtitle and evidence heading, a person reads in natural order (given name, then surname), because it sits in a sentence.
   - Today the row's own name is natural already (`ObservationValueDisplay` uses `.natural`).
   - The **record's** title comes from `.handle(id:ref:title:)`, built by `PromoteFlow.conclusionTitle` with `PersonHeaderDisplay.title(person, style: .sorted, …)` for the record menu (`PromoteFlow.swift:690`).
2. The "Additional comments or arguments" text area fills the vertical space of its container: the height of the Status and Confidence column beside it. Today it is a few lines tall (`PromoteEvidenceSheet.swift:135`).

### Fix plan 19
**F-42 — The sheet's sentences read a person's record title in natural order** (from the user)
Files: `macos/App/Features/Promote/{PromoteFlow,PromoteEvidenceSheet,PromoteEvidenceDisplay}.swift`, tests.
Failing test first: a person row whose record header has given "John", surname "Smith" and a sorted order. The record menu still reads "Smith, John" (or the sorted form today's test pins). The sheet's title, subtitle and evidence heading read "John Smith".
Do:
- Each record alternative carries both forms: `title` (sorted, for the menu, unchanged) and `readingTitle` (natural: `PersonHeaderDisplay.title(person, style: .natural, orders:)`). Events and places use the same string for both.
- Carry `readingTitle` on `.handle` (or look it up from the row's menu by id, with no second copy of the title).
- The sheet's title, subtitle and heading read `readingTitle`.
- The record menu and the page row are unchanged.
Check: test passes; scripts/check-pr.sh PASS.

**F-43 — The argument box fills its container's height** (from the user)
Files: `macos/App/Features/Promote/PromoteEvidenceSheet.swift`; `macos/App/DesignSystem/Components/TextArea/PVTextArea.swift` and `macos/App/DesignSystem/README.md` only if the kit needs an option.
Do:
- The claim fields' `HStack` sizes to its tallest column. The argument `PVField` and its `PVTextArea` stretch to that height, top to bottom of the Status + Confidence column, using `.frame(maxHeight: .infinity)` within a row that is `.fixedSize(horizontal: false, vertical: true)`, or the kit's equivalent.
- If `PVTextArea` sizes itself to its content and cannot fill, add one documented option (e.g. `fillsHeight: Bool`), not a sheet-local copy.
- Text still scrolls inside it when longer than the box.
Check: scripts/check-pr.sh PASS; say how you verified the layout (no UI test target).

## Round 19 — fix
Head `1170f33b` (`279cde5f` F-42 plus the sheet's F-43 lines, `1170f33b` `PVTextArea.fillsHeight`). Gate PASS (orchestrator re-ran).
- **F-42:** `aPersonRecordReadsNaturallyInTheSheetAndSortedInTheMenu` failed first. `Alternative.readingTitle` is natural; `Row.readingTitle` looks it up by id and falls back to the target's title. The menu and the row are unchanged.
- **F-43:** `PVTextArea(fillsHeight:)` (default false); the claim row is `fixedSize(vertical)` and the argument field `maxHeight: .infinity`. Not verified on screen (no UI target).
- No new warning in changed files (fresh derived data).

## Result (after round 19)
Gate PASS · CI see PR · open: none · manual check §10 not run · F-43 layout to be checked by eye

## Round 20 — user request (2026-10-23)
"Show N records" sat 10 pt right of the Source titles below it: the ghost `PVDisclosureButton` keeps `PVControlSize.sm.horizontalPadding` inside its label. At the user's request this was a quick fix without a review round, made by the orchestrator: `.padding(.leading, -PVControlSize.sm.horizontalPadding)` on the disclosure in `EvidenceLine.sourceColumn`. Gate PASS. Not checked on screen.
