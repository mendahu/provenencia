# Ship log — S13-35 The person card joins the design system

**Plan:** [`S13-35-person-card-joins-the-design-system.md`](S13-35-person-card-joins-the-design-system.md) · **Branch:** `s13-35-person-card` · **PR:** #690
**State:** ready

Base: `main` @ `a0275471` (plan written against `04a2c99`). Plan Status was "Blocked — on S13-34"; S13-34 landed in #689 (`a0275471`), review log complete, Open questions None, so the loop starts. STEP-0 re-checks S13-34's landed names.

## Round 0 — implement
Head `9a8116f9`. Gate PASS (orchestrator run against `a0275471`: all nine checks). PR #690 (draft).
Report deviations: (1) compact "abt" asserts "abt 1790 – 1851" (S13-34's no-dot wording), not "abt."; (2) `PersonDetailContent.badge(_:)` is a static function taking the treatment, `badge` removed from `PVGenderTreatment`; (3) the shell detects an absent trailing view with `Trailing.self is EmptyView.Type`; (4) the regular lines slot applies no line limit (as the old tile); only the compact years line reads `linesLineLimit`; (5) catalog keys `personCard.thisPerson`, `thisPersonSpoken`, `noLikeness`; `conclusions.family.thisPerson` stays; (6) commit 1 leaves old `PersonTile.swift`/`PersonGenderTreatment.swift` so each commit builds; (7) Landed note lists 1–4; (8) Manual check not run; `#Preview`s added, not rendered. Sweep leftovers only in spike history and the S13-D4 brief (S13-13 owns it).

## Round 1 — review
Reviewers: CON (Sonnet), MNT, UX. Skipped: SEC+PRF (Swift-only view refactor; no `core/**`, FFI, migrations, catalog-loaded sizes or loops over data the app did not write; `GenealogyStore.swift` is a comment change, `PersonLifeDisplay` adds two formatters, Family layout changes are type renames).

| ID | Lens | Severity | Title | Decision | Evidence / reason |
| --- | --- | --- | --- | --- | --- |
| MNT-1 | MNT | major | `LifeLine.text` is a third hand-written date · place join | Fix now → F-1 (same-pass cleanup) | verified: `rg 'lineJoin\(' macos/App` → `PersonLifeDisplay.swift:26` (`line(_:)`), `:44` (`LifeLine.text`, new), `EventsListView.swift:90` (`EventsList.line`), each with its own missing-half switch |
| UX-3 (MNT out-of-lens) | UX | minor | `LifeLine.text` builds "b. 3 May 1790" by string interpolation; word order not localizable | Fix now → F-1 | verified: `PersonLifeDisplay.swift:48` `"\(abbreviation) \(rest)"`; new code in this PR |
| CON-1, MNT-2 | CON, MNT | minor | Person card always draws the title-row `HStack`; D-1's `EmptyView` branch and Risk-row mitigation never fire | Fix now → F-2 | verified: `PVPersonCard.swift:108` always passes `trailing: { if … { PVBadge } }` (type `Optional<PVBadge>`), `PVTileCard.swift:91` reflection is then always true; the `where Trailing == EmptyView` init has no caller. Fixed within D-1 (card picks the trailing-less init), not by MNT-2's "drop the branch", which contradicts D-1 |
| UX-1, CON-2, MNT-4 | UX, CON, MNT | minor | Translator comment on `conclusions.family.thisPerson` still says it is the badge | Fix now → F-3 | verified: `rg familyThisPerson macos/App` → definition + `PersonFamilyContent.swift:159` only |
| CON-3 | CON | minor | `LifeLine.text` place-only / empty branches untested | Fix now → F-1 (test) | `compactYearsAndHelp` has no place-only birth |
| MNT-5 | MNT | minor | Spike docs say "until S13-35" and S13-13 cites the old working-rule title | Fix now → F-4 | verified: `deployment-plan.md:111`, `person-detail.md:13`, `S13-13-timeline-section.md:15` |
| MNT-3 | MNT | minor | Composer row pairs kit Return with a hand-written Space handler under the same guard | Rejected | D-2 binds: "keeps its own `.focused` and Space handler"; one caller pairs the two, no hand-rolled list to collapse |
| CON-5 (MNT out-of-lens) | CON | minor | `add-ui-component` checklist asks for a test of each promised behavior; `PVTileCard` Return/focus/hover and `pvActivatesOnReturn` have none | Follow-up | no Swift test in the repo drives a key press (`rg -l onKeyPress macos/ProvenenciaTests` empty; `PVConfirmReturnTests` holds a tone property, not key handling). A unit seam for SwiftUI key handling is kit-wide work; manual checks 2 and 4 cover it here |

Nits (logged): `UX-2 · PVFocusRingButtonStyle.swift:19 · state manual check 4's result (ring follows .focused(…, equals:), incl. focusedID after "Use as likeness") in the PR body` · `CON-4 / MNT-7 · PVGenderTreatment.swift:4 · rewrap the edited doc line` · `MNT-6 · PersonDetailContent.swift:15,19 · rename static badge(_:) to genderBadge(_:)`

### Fix plan 1
Order: the major, then minors riding along. Fold the two nits on lines this round touches (CON-4/MNT-7 rewrap, MNT-6 rename) in if trivial.

**F-1 — One rule for the date · place text** (from MNT-1, UX-3, CON-3)
Files: `macos/App/DesignSystem/Components/DatePlaceLine/PVDatePlaceLine.swift` (wherever `PVDatePlaceLine` lives), `macos/App/Platform/PersonLifeDisplay.swift`, `macos/App/Features/Conclusions/EventsListView.swift`, `macos/App/Platform/L10n.swift`, `macos/App/Resources/Localizable.xcstrings`, `macos/ProvenenciaTests/PVPersonCardTests.swift` (and the existing `PVDatePlaceLine`/Events tests). Skills: `add-localized-string`, `add-swift-test`.
Failing test first: add to `PVPersonCardTests/compactYearsAndHelp` a place-only birth (Bremen, no date) → `helpText` "Anna Weber\nb. Bremen" (with a death of 1851 → "…\nb. Bremen\nd. 1851"); and a new `PVDatePlaceLine` test of the shared join: (date, place) → "3 May 1790 · Bremen"; (date, nil) and (date, "") → date; (nil, place) and ("", place) → place; (nil, nil) → "".
Do: add `nonisolated static func text(date: String?, place: String?) -> String` on `PVDatePlaceLine` beside `spokenDate` / `spokenPlace`, treating nil and "" alike as a missing half and joining with `L10n.Conclusions.lineJoin`. Move all three copies onto it: `PersonLifeDisplay.line(_:)` (its `dates`/`places` switch), `LifeLine.text`, `EventsList.line(_ content:)`. For `LifeLine.text`, join the abbreviation and the rest through a new localized format (e.g. `L10n.Conclusions.lifeLineText(abbreviation:rest:)`, key `conclusions.lifeLine.text`, default `"%1$@ %2$@"`, translator comment naming the compact card's tooltip line "b. 3 May 1790 · Bremen") instead of string interpolation; the abbreviation alone when the rest is empty.
Check: the grep `rg -n 'lineJoin\(' macos/App` prints only the `L10n.swift` definition and the one call inside `PVDatePlaceLine.text`; `rg -n '\\\(abbreviation\)' macos/App` prints nothing; the new and existing tests pass (`PersonLifeDisplay`, Events list, `PVPersonCardTests`); scripts/check-pr.sh PASS.

**F-2 — The person card takes the trailing-less shell when it has no badge** (from CON-1, MNT-2)
Files: `macos/App/DesignSystem/Recipes/PersonCard/PVPersonCard.swift`.
Failing test first: none (layout; held by manual check 1 and D-1).
Do: in `PVPersonCard.body`, branch on `content.isFocus && size == .regular`: with the badge, the four-slot `PVTileCard` init with `trailing: { PVBadge(L10n.PersonCard.thisPerson) }`; without it, the `where Trailing == EmptyView` init. Factor the shared thumbnail/title/lines builders so neither branch repeats them. Keep D-1's `PVTileCard` as written (the reflection branch and the extension stay).
Check: `rg -n 'trailing:' macos/App/DesignSystem/Recipes/PersonCard/PVPersonCard.swift` shows one call whose closure holds only `PVBadge`, no `if`; T-1, T-3, T-5 pass; scripts/check-pr.sh PASS.

**F-3 — Translator comment on `conclusions.family.thisPerson`** (from UX-1, CON-2, MNT-4)
Files: `macos/App/Platform/L10n.swift` (~9230), `macos/App/Resources/Localizable.xcstrings` (~18250).
Do: reword the comment in both to "The label of the focus alone in the siblings group of the Family section."
Check: `rg -n 'Badge on the focus tile' macos/App` prints nothing; localizable PASS.

**F-4 — Spike docs name the move as done** (from MNT-5)
Files: `docs/deployment-plan/spike-13/deployment-plan.md:111`, `docs/deployment-plan/spike-13/person-detail.md:13`, `docs/deployment-plan/spike-13/prs/S13-13-timeline-section.md:15`.
Do: drop the "`PersonGenderTreatment` until S13-35 moves it into the design system" clauses; rename the working rule in S13-13's Context line to "Colors reuse `PVGenderTreatment`". Leave the S13-D4 brief alone (S13-13 owns it). Add F-1's shared join to the plan's Landed note as a difference from the plan.
Check: `rg -n 'until S13-35|Colors reuse .PersonGenderTreatment' docs/deployment-plan/spike-13` prints nothing; plan-docs PASS.

## Round 1 — fix
Head `c8490e8e`. Gate PASS (orchestrator run against `a0275471`: all nine checks). F-1 to F-4 done (`730f58b1`, `1981380f`, `c8490e8e`).
Report deviations: F-1's new `PVDatePlaceLine.text` test failed by not compiling before the helper; the CON-3 place-only case was added as coverage, not run red first. F-4's grep still hits the S13-D3/S13-D4 briefs, the landed S13-02/S13-12 plans and this plan's own D-10 and Rules rows: accepted, they are history or S13-13's brief. Nits MNT-6 and CON-4/MNT-7 not applied (files untouched this round); stay Logged.

Minors read by the orchestrator in the fix commits:
- CON-1/MNT-2 (F-2): **Resolved.** `PVPersonCard.body` takes the four-slot init only with the badge, and its `trailing:` holds `PVBadge` alone; otherwise the `Trailing == EmptyView` init. The builders are shared as private properties.
- UX-1/CON-2/MNT-4 (F-3): **Resolved.** The comment reads "The label of the focus alone in the siblings group of the Family section." in both files.
- UX-3 (F-1): **Resolved.** `conclusions.lifeLine.text` with a translator comment; no `\(abbreviation)` interpolation left.
- CON-3 (F-1): **Resolved.** Place-only birth case in `compactYearsAndHelp`.
- MNT-5 (F-4): **Resolved** for the live docs (map, `person-detail.md`, S13-13's Context line).

## Round 2 — review
Reviewers: MNT (scope `9a8116f9...c8490e8e`; the only lens with a Fix-now major). CON and UX had minors only, read above.
MNT-1: **Resolved** (`PVDatePlaceLine.text(date:place:)` at `PVDatePlaceLine.swift:57-65`; `rg 'lineJoin\(' macos/App` → definition + that one call; six nil/empty cases in `PersonDetailTests.datePlaceLineTextJoinsWhatIsRecorded`). New findings: none. Verdict: Clean. No Fix plan 2.

## Result
Gate PASS (`c8490e8e`) · CI green (localizable, test; run 41537748130) · open: none · follow-ups: CON-5 (a unit seam for SwiftUI key handling so kit Return/focus behaviours can be held by a test) · to user: — · CON miss: none
`origin/main` unchanged since the branch was cut (`a0275471`); no merge needed.
