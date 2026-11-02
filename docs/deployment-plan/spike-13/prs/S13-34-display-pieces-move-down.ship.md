# Ship log — S13-34 Display pieces move down a layer

**Plan:** [`S13-34-display-pieces-move-down.md`](S13-34-display-pieces-move-down.md) · **Branch:** `s13-34-display-pieces-move-down` · **PR:** #689
**State:** ready

Base: `main` @ `f21ee599` (plan written against `04a2c99`).

## Round 0 — implement
Head `5c2b2a79` (local, not pushed; no PR). Gate not run. STEP-1 to STEP-6 committed, each step's named suites passing per the report; STEP-7 and STEP-8 not done.

Report deviations: (1) `PVTitleText.italicizingEmphasis` kept internal, not private, because `PersonTitleDisplayTests.swift:122` calls it; (2) `.year` wording asserts the app's real "abt 1790" / "bef 1790" (no dot), not AC-2.4's "abt."; (3) a `.year` range with a year on one side only shows "" (plan silent); (4) STEP-2 commit carries a one-line interim `PVBadge.moreCount` edit in `EventsListView` so it compiles; (5) `EventSecondaryLine` stays as a thin wrapper keeping the "neither half → no second line" guard; (6) new `designSystem.*` catalog entries placed before `designSystem.checkbox.locked`.

**Stopped on:** STEP-7, D-12 / stop condition "kit check finds hits beyond `WorkspaceLocation` at the base". The `^\s*…` harvest picks up **nested** feature types with generic names, giving 15 false hits in kit files: `Group` (SwiftUI's, vs nested `struct Group` in `PersonFamilyLayout.swift`, `PersonFamilyContent.swift`), `Action` (`CatalogLastUpdated.swift:124`), `Draft` (Metadata, SourceTypes, Sources), `Family` (`PersonGenderTreatment.swift:8`), `Open` (`SourcesGraphButton.swift:19`), `Field` (`PromoteView.swift:481`). Draft check saved at the session scratchpad `step7-kit-check-draft.patch`. Taken to the user; the D-12 fix goes back through `plan-pr`.

User answer (2026-11-02): harvest top-level declarations only. D-12 amended in commit `7a7e086d` (plan + review log A-1). Implementer resumed at STEP-7.

## Round 0 — implement (resumed)
Head `9f215942`. Gate PASS (orchestrator run against `f21ee599`: all nine checks). PR #689 (draft).
Report deviations (new): (7) `PVDatePlaceLine.spokenDate`/`spokenPlace` and `PVTitleText.italicizingEmphasis` marked `nonisolated` (swift-warnings; `PVClipboardButton` precedent); (8) `plan-docs` fails while the plan is Planned (§ 4 cites moved paths), so the gate passes only after the landing commit; (9) a local commit briefly held the ship log, reset before push. AC-2.2 one-off `rg` (453 names over 9 moved Platform files) empty. § 10 manual check not run.

## Round 1 — review
Reviewers: CON (Sonnet), MNT, UX, SEC+PRF. UX: `macos/App/**`, `Localizable.xcstrings`. SEC+PRF: `api/ffi/**` (one test constant) and `macos/App/Features/**` models (`ReconciledValueRowModel`), dispatched on the path match.

| ID | Lens | Severity | Title | Decision | Evidence / reason |
| --- | --- | --- | --- | --- | --- |
| MNT-1 / CON-2 / UX (outside lens) | MNT, CON, UX | major (MNT rated blocker; re-rated: duplicated join, no wrong output today) | `EventSecondaryLine` adds a private copy of `EventSecondaryDisplay.line`'s date · place join | Fix now → F-1 | verified `EventsListView.swift:83-88` vs `:113-117`; D-5 names `EventSecondaryDisplay.line(…)`'s text as the label |
| UX-1 | UX | minor | Events row's date now draws in `textPrimary`, not the row's muted ink; § 1 says "the date … as before" | Fix now → F-2 | verified: `PVDatePlaceLine.swift` sets the date `.foregroundStyle(PVColor.textPrimary)`; base `EventSecondaryLine` set none (inherits `PVList`'s `textMuted`, `PVList.swift:155`). The user approved only the place's style; the plan's § 1 binds the date |
| MNT-2 | MNT | minor | `WorkspaceLocation` kit exemption filed under "Apple frameworks" in the doc-name allow file; every doc-only name also exempts the kit check | Fix now → F-3 | verified allow file header and `features -= allowed()` |
| MNT-3 | MNT | minor | `Missing`'s orphaned doc text merged into `lifeLines`' doc | Fix now → F-4 | verified `PersonLifeDisplay.swift:40-44` |
| CON-1 | CON | minor | Two T-3 bridge tests would pass on the base code | Fix now → F-5 | tests this PR adds; T-3 asks for the property-key assertion |
| SEC+PRF | SEC, PRF | — | No findings | — | — |

Nits (logged): `CON-3 · plan AC-2.4 · "abt." vs the app's "abt"; Landed note covers it` · `MNT-4 · PersonGenderTreatment.swift:74 · redundant !text.isEmpty` · `MNT-5 · check-guidance-names.py · "//" strip cuts at "https://" literals; say so` · `MNT-6 · PVTitleText.swift:118 · italicizingEmphasis nonisolated only for tests; commit subject overstates` · `UX-2 · life-span year form reads "bet 1790 and 1795" for a from/to range; check S13-35 tooltip copy`

### Fix plan 1

**F-1 — One date · place join for the Events row** (from MNT-1, CON-2)
Files: `macos/App/Features/Conclusions/EventsListView.swift`, `macos/ProvenenciaTests/PersonsListTests.swift` if a call site changes.
Do: split `EventSecondaryDisplay.line(_:locale:)` into `static func line(_ content: Content) -> String` (the join) and `line(_:locale:)` calling `line(content(summary, locale: locale))`. `EventSecondaryLine` takes the `Content` that `EventsList.secondary` already builds (or keeps its three fields and builds one) and passes `EventSecondaryDisplay.line(content)` as `accessibilityLabel`. Delete the private `line`.
Check: `rg -n "lineJoin" macos/App/Features/Conclusions/EventsListView.swift` prints exactly one line (inside `EventSecondaryDisplay.line(_ content:)`); `eventsRowSecondLine` and `PersonsListTests` pass; scripts/check-pr.sh PASS.

**F-2 — The Events row's date keeps its muted ink** (from UX-1)
Files: `macos/App/DesignSystem/Components/DatePlaceLine/PVDatePlaceLine.swift`, `EventsListView.swift`.
Do: add `dateTone: Color = PVColor.textPrimary` to `PVDatePlaceLine.init` beside `dateSize` (default keeps the header, Family tiles and Event page unchanged); the date `Text` uses it. `EventSecondaryLine` passes `dateTone: PVColor.textMuted`, matching the base row (which inherited `PVList`'s `textMuted`). Note the added parameter in the plan's Landed note as a deviation from D-5's signature, citing § 1 "the date … as before".
Check: `rg -n "dateTone" macos/App` shows the init, the date Text, and the Events call only; scripts/check-pr.sh PASS.

**F-3 — The kit check's allow entries live in their own section** (from MNT-2)
Files: `scripts/check-guidance-names.allow`, `scripts/check-guidance-names.py`.
Do: update the allow file's header to describe both uses. Move `WorkspaceLocation` under a new final heading `# Kit check: feature types design-system code may name`. `kit_names_feature_types` subtracts only the names under that heading (parse sections in the script; the guidance check keeps reading every entry as today).
Check: `python3 scripts/check-guidance-names.py` prints OK; temporarily moving `WorkspaceLocation` above the new heading makes the kit check fail on `PVDeleteImpact.swift` (revert after); scripts/check-pr.sh PASS.

**F-4 — `lifeLines`' doc comment reads summary first** (from MNT-3)
Files: `macos/App/Platform/PersonLifeDisplay.swift`.
Do: put "The *b.* and *d.* lines of a Person." first, then a blank `///` line, then the `.stated` / `.omitted` sentences reworded as the description of the `missing:` parameter.
Check: read lines 36-48; scripts/check-pr.sh PASS.

**F-5 — The bridge tests prove D-3** (from CON-1)
Files: `macos/ProvenenciaTests/EvidenceBridgeEdgeSummaryTests.swift`.
Failing test first: not applicable (strengthens tests over shipped behavior); instead show each new assertion fails when the D-3 change is reverted locally, then restore.
Do: in `everyBridgeKindReadsItsTermFromTheSeededProperty`, also assert each bridge fixture observation's `propertyKey == ConnectionSentenceRegistry.termProperty(for: kind)`. In `aRelabelledProductRoleReadsWithTheAppsName`, assert the exact sentence the summary yields, and use a fixture where the base `termLabel` path would have read differently from `ObservationValueDisplay` if one exists; if none does, say so in the report.
Check: both tests pass; scripts/check-pr.sh PASS.

## Round 1 — fix
Head `a1f8d168`. Gate PASS (orchestrator run, all nine checks). F-1–F-5 done.
- UX-1 (F-2): **Resolved.** `dateTone` on `PVDatePlaceLine` (default `textPrimary`); the Events row passes `textMuted`; Landed note deviation (7).
- CON-1 (F-5): **Resolved in part.** The exact sentence is pinned. The per-kind `propertyKey` assertion builds its observation from the expected key, so it is a tautology; the implementer showed no fixture can tell D-3 from the base path (both read the same key when the fixture is correct). Accepted: D-3's value is the single path, not a different output.
- MNT-1, MNT-2, MNT-3 → round 2 (MNT had a Fix-now major).

## Round 2 — review
Reviewers: MNT (scope `9f215942...a1f8d168`). CON and UX had only minors (checked above); SEC+PRF had none.

| ID | Lens | Severity | Title | Decision | Evidence / reason |
| --- | --- | --- | --- | --- | --- |
| MNT-1 | MNT | major | Duplicate date · place join | Resolved | reviewer: one `lineJoin` at `EventsListView.swift:90`; only caller passes `Content` |

No new findings. MNT-2 **Resolved** (orchestrator read: allow-file header describes both uses; `kit_allowed()` reads only the `# Kit check:` section, guidance check keeps `allowed()`). MNT-3 **Resolved** (`PersonLifeDisplay.swift:40-46` summary first, `missing` described after).

## Result
Gate PASS · CI green (both `test` jobs pass; `localizable` skipped) · open: none · follow-ups: none · to user: — (D-12 STEP-7 stop answered: top-level only) · CON miss: none
