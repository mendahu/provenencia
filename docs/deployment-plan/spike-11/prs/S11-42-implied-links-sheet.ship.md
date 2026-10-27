# Ship log — S11-42 Implied links on the evidence sheet

**Plan:** [`S11-42-implied-links-sheet.md`](S11-42-implied-links-sheet.md) · **Branch:** `s11-42-implied-links-sheet` · **PR:** #585
**State:** ready

## Round 0 — implement
Head `0b771b50`. Gate PASS (orchestrator run). PR #585 (draft). Report deviations: `PromoteEvidenceDisplay.wording` has an unreachable `.direct` arm (exhaustive switch); `EvidenceBoard.Pivot` declares its memberwise init itself; `ConclusionHeaderDisplay.swift` added to `project.pbxproj`; both code deviations in the plan's Landed note. AC-1.12 and §10 manual check left to the user.

## Round 1 — review
Reviewers: CON (Sonnet), MNT, UX, SEC+PRF. Skipped: none (UX: `macos/App/**` and `Localizable.xcstrings`; PRF: `macos/App/Features/**` models — `PromoteModel`, `EvidenceGraphSnapshot`).

| ID | Lens | Severity | Title | Decision | Evidence / reason |
| --- | --- | --- | --- | --- | --- |
| MNT-1 | MNT | major (rated blocker) | `direct(term:)` is a second per-term wording table beside `L10n.PropertyTerm.Wording` | To user → Fix now → F-1 (user, 2026-10-27: move to Wording) | verified: `PromoteEvidenceDisplay.swift:246-256` switches on `SeededRelationshipTypeTerm` while `L10n.PropertyTerm.wording` (`L10n.swift:2642`) already holds each term's words and `add-seeded-vocabulary` names only that table. Re-rated major: a new term falls back to "recorded directly", no wrong data. Contradicts D-4's private switch |
| UX-1 | UX | minor | "on %@" reads wrongly for year-only, qualified, ranged dates; a date phrase's "·" collides with the caption's side separator | To user → Fix now → F-4 (user, 2026-10-27: "dated %@", caption date without its phrase) | verified: `DateRowDisplay.civil` returns a bare year; `DateRowDisplay.swift:109` joins a phrase with "·"; `implied` passes the row line into `recordedOn`. Copy is set by §6 Strings ("on %@") |
| MNT-2 | MNT | minor | `wording(_:)` has a dead `.direct` arm and the caller re-checks `.direct` | Fix now → F-2 | verified: `PromoteEvidenceDisplay.swift:197-199, 228-230`; file this PR adds |
| CON-1 | CON | minor | Nothing checks that `PromoteModel.subjectFacts` sets `date` | Fix now → F-3 | verified: `rg subjectFacts macos/ProvenenciaTests` empty; `PromoteModel.swift:484` is code this PR adds |
| MNT-3 | MNT | minor | `PromoteFlow.item` rebuilds a `CatalogConclusionHeader` that `GoStore` split into three optionals | Follow-up | Reshape of `GoStore`/`FakeStore` mapping (plan Out); the split predates this PR |
| SEC/PRF | SEC+PRF | — | No findings | — | — |

Nits (logged): `MNT-4 · PromoteEvidenceBoard.swift:232-234 · Pivot(row:) init would mirror init?(header:orders:) and drop the hand-written memberwise init` · `UX-2 · L10n.swift recordedDirectOther · "recorded directly" for a sibling; "recorded as siblings" would match the other direct phrases`

**User answers (2026-10-27):** MNT-1 — move the direct words onto `L10n.PropertyTerm.Wording` (supersedes D-4's private switch). UX-1 — `promote.sheet.recorded.on` becomes "dated %@" and the caption's pivot date drops its recorded phrase (supersedes §6 "on %@").

### Fix plan 1
**F-1 — Direct-side words live on `L10n.PropertyTerm.Wording`** (from MNT-1; user answer supersedes D-4's private switch)
Files: `macos/App/Platform/L10n.swift`, `macos/App/Features/Promote/PromoteEvidenceDisplay.swift`, `.agents/skills/add-seeded-vocabulary/SKILL.md`, the plan (D-4 Landed note, Next adds). Skill: `add-localized-string`, `add-seeded-vocabulary`.
Failing test first: none new — T-1 `recordedNamesHowEachSideRecordsIt` and T-2 must pass with no expected value changed (same keys, same text).
Do: add `var recordedDirect: LocalizedStringResource?` to `L10n.PropertyTerm.Wording` with a doc comment in the style of `share`. Move the four `recordedDirect*` resources so `relationshipSpouse` gets the spouses one, `relationshipParent` and `relationshipChild` share one `static let`, `relationshipGrandparent` and `relationshipGrandchild` share one; keep the existing keys and default values. `recordedDirectOther` stays in `L10n.Promote` as the fallback. `direct(term:)` becomes a lookup: `term.flatMap { L10n.PropertyTerm.wording(propertyKey: SeededPropertyKey.relationshipType, termKey: $0)?.recordedDirect }.map(L10n.string) ?? L10n.string(L10n.Promote.recordedDirectOther)`. Update the `add-seeded-vocabulary` step that lists a relationship type's words to name `recordedDirect`, and `add-match-rule` step 6 / the plan's Next adds line if they name `direct(term:)`.
Check: `rg -n "case SeededRelationshipTypeTerm\." macos/App` prints only `L10n.swift` lines; T-1, T-2 pass; scripts/check-pr.sh PASS.

**F-2 — `wording(_:)` returns nil for `.direct`** (from MNT-2)
Files: `macos/App/Features/Promote/PromoteEvidenceDisplay.swift`, the plan's Landed note.
Failing test first: none — refactor; T-1 and T-2 unchanged.
Do: `wording(_ pattern:) -> PatternWording?`, `case .direct: nil` (switch stays exhaustive). `side` becomes `wording(pattern).map { implied($0, pivot) } ?? direct(term: term)`. Remove the Landed note line about the dead `.direct` arm.
Check: `rg -n "Never asked for" macos/App` prints nothing; scripts/check-pr.sh PASS.

**F-3 — Test that `subjectFacts` carries an event's date to its row** (from CON-1)
Files: `macos/ProvenenciaTests/PromoteModelTests.swift` (or `PromoteFlowTests.swift` if its fixtures load a page through `PromoteModel`). Skill: `add-swift-test`.
Failing test first: `aPlacedEventsRowCarriesItsDate` — load a page with a placed `.event` Subject holding a positive `date` Observation through `PromoteModel`; assert `model.flow.row(id)?.date == DateRowDisplay.line(date: thatDate)`. Prove it red by deleting `date: placed.dateLine` at `PromoteModel.swift:484` locally, then restore.
Do: test only.
Check: the test passes; scripts/check-pr.sh PASS.

**F-4 — "dated %@", and the caption's pivot date without its phrase** (from UX-1; user answer supersedes §6 "on %@")
Files: `macos/App/Platform/L10n.swift`, `macos/App/Resources/Localizable.xcstrings`, `macos/App/Features/Dates/DateRowDisplay.swift`, `macos/App/Features/Workspace/Session/EvidenceGraphSnapshot.swift`, `macos/App/Features/Promote/PromoteEvidenceBoard.swift` (and `EventSecondaryDisplay` if its `.date` is where the record side's phrase is added), tests, the plan (§6 Strings row, Landed note). Skill: `add-localized-string`.
Failing test first: T-1 case expectations move from "on 12 May 1842" to "dated 12 May 1842"; add a T-1 or T-2 case whose pivot date carries a phrase (e.g. date 1842 with phrase "in the spring") on each side, expecting "…dated 1842" with no "in the spring" and no "·" inside a side.
Do: `promote.sheet.recorded.on` default "dated %@" (keep the key; update its comment). Give `DateRowDisplay.line` a defaulted parameter that leaves the recorded phrase off (default keeps today's output everywhere else). The caption's two pivot dates use it: the Source side through what `Pivot` reads from the page row, the record side through `Pivot.init?(header:orders:)`. Do not change what the Events list or Promote rows show elsewhere — if `Row.date` is read anywhere besides the pivot, give the pivot its own phrase-free date rather than changing `Row.date`'s meaning. Update §6's row and the Landed note.
Check: T-1/T-2 pass with the new cases; `python3 scripts/check-localizable-xcstrings.py` OK; scripts/check-pr.sh PASS.

## Round 1 — fix
Head `952f14d6`. Gate PASS (orchestrator run). F-1 to F-4 done. Report deviations: F-4's phrase case is tested at `dateLine`, `Pivot(header:)` and `DateRowDisplayTests` rather than in T-1 (`Pivot.date` is already a string); `dateLine` itself is now phrase-free since only the pivot reads `Row.date` (verified: `rg dateLine macos/App` — Events list builds its own line in `EventsListView.swift:75`); F-1's commit carries F-4's "dated" default; D-4/D-10 rows left as planned, differences in the Landed note.

## Round 2 — review
Reviewers: MNT (MNT-1 major, Fix now). Not sent: CON, UX (Fix-now minors only), SEC+PRF (no findings).

| ID | Lens | Severity | Title | Decision | Evidence / reason |
| --- | --- | --- | --- | --- | --- |
| MNT-1 | MNT | major | Direct-side words on `L10n.PropertyTerm.Wording` | Resolved | `direct(term:)` is a lookup; `rg "case SeededRelationshipTypeTerm\." macos/App` finds only the `wording` table; skill and Next adds updated |

No new findings. Round-1 minors checked by the orchestrator: MNT-2 **Resolved** (`wording` returns `PatternWording?`, no `.direct` arm); CON-1 **Resolved** (`PromoteModelTests.aPlacedEventsRowCarriesItsDate`, red with the line removed); UX-1 **Resolved** (`promote.sheet.recorded.on` = "dated %@"; `DateRowDisplay.line(includesPhrase: false)` at both pivot dates).

## Finish
`origin/main` unchanged since the branch was cut (`731e4239`); no merge. Orchestrator docs edit with this log: "the marriage on 12 May 1842" / "on <date>" → "dated" in `deployment-plan.md`, `kinship-shapes.md` and the plan (ACs, §10), to match UX-1's copy.

## Result
Gate PASS · CI pending · open: none · follow-ups: MNT-3 · to user: MNT-1, UX-1 (answered 2026-10-27) · CON miss: none
