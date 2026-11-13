# Ship log — S15-35 The Artifact list shows each Artifact's copy details

**Plan:** [`S15-35-artifact-list-rows.md`](S15-35-artifact-list-rows.md) · **Branch:** `s15-35-artifact-list-rows` · **PR:** #824
**State:** ready

## Round 0 — implement
Head `e771ec007`. Gate PASS (orchestrator run, against `d18918601`). Report deviations: STEP-1 and STEP-2 in one commit; `artifactGraphsSpoken` and the `restingCard…` → `restingRow…` rename landed in the STEP-4 commit so every commit builds; § 10 manual check not run (no GUI session); no `#Preview` on the row (D-7).

## Round 1 — review
Reviewers: CON, MNT, UX, SEC+PRF. UX: the diff touches `macos/App/**` and `Localizable.xcstrings`. SEC+PRF: no security path matches, but the diff adds list code (the Artifacts `LazyVStack`, `graphLines` derived per row from catalog data) and a layout component (`PVFlowLayout.arrange`), which the performance paths cover.

| ID | Lens | Severity | Title | Decision | Evidence / reason |
| --- | --- | --- | --- | --- | --- |
| CON-1, UX-1 | CON, UX | blocker | An Artifact with no copy details loses its column; its text runs under the other rows' details | Fix now → F-1 | verified: `ArtifactCopyDetailsView.body` is a bare `if !shown.isEmpty` with no else, so the row's flexible `.frame` has no view to size; UX measured the text column at 536 → 866pt in a hosted probe. Breaks AC-1.2 and D-7's "still takes the column" |
| CON-2 | CON | major | § 10 manual check not run; most ACs rest on it | To user (finish report) | Not a code defect: the user runs § 10 before merge, as ship-pr step 8 hands it over. Report flags steps 1, 3, 5, 6, 7 and UX-5 |
| CON-3 | CON | minor | T-1 never reaches D-3's "still wider than the row" clamp | Fix now → F-2 | verified: the test's `fit` returns `min(widths[i], offered)`, so the clamp at `PVFlowLayout.swift:148-151` never runs |
| MNT-1 | MNT | minor | Row buttons repeat the calls `perform(_:)` makes | Fix now → F-3 | verified at `SourcePageArtifactRow.swift`; the menu and buttons would drift when S15-25 adds Move |
| MNT-2 | MNT | minor | Menu takes both `title` and `verbatimTitle`; one silently wins | To user (finish report) | D-13 specifies `verbatimTitle: String? = nil` *beside* `title:`; the finding contradicts a Decision but follows the kit's `PVEmptyState` precedent better |
| MNT-3 | MNT | minor | `PVTag` uses literals 6 and 8 where `PVSpacing.space3/space4` exist | Fix now → F-4 | verified: `PVSpacing.swift:12-13`; Components use tokens |
| MNT-4 | MNT | minor | `SourceFormLoadArea` doc comment still says "for a card" | Fix now → F-4 | stale comment in a block this PR edited; only `ArtifactSheet` passes `compact: true` |
| UX-2 | UX | minor | Copy-details field name truncates at 96pt with no tooltip | Fix now → F-5 | the value beside it has `.help`; label column is new in this PR (D-9) |
| UX-3 | UX | minor | Truncated / "+n" graph names have no tooltip | Fix now → F-5 | `PersonSourcesSection.swift:460` precedent; `all` is in scope |
| UX-4 | UX | minor | Person page type Tag now truncates with no tooltip | Fix now → F-5 | truncation is new with `PVTag` (D-1); add `.help(label)` in `typeCell`, not in `PVTag` |

SEC+PRF: no findings (Clean).

Nits (logged): `MNT-5 · SourceArtifactsSection.swift:121-149 · names[id] == nil already means .none` · `MNT-6 · PVFlowLayout.swift:83-86,105-108 · shared fit closure helper` · `MNT-7 · CachedThumbnail.swift:35-41 · fold tileBody into body` · `UX-5 · SourcePageArtifactRow.swift:69-83 · check the actions rotor in manual step 7` · `CON-4 · SourcePageIdentityHeader.swift:218 · .layoutPriority(-1) on the inner date Text is a no-op (D-4 placed it there)`

### Fix plan 1

**F-1 — The copy-details column always exists** (from CON-1, UX-1)
Files: `macos/App/Features/Sources/ArtifactCopyDetailsView.swift`, `macos/App/Features/Sources/SourcePageArtifactRow.swift`.
Failing test first: none reachable (SwiftUI layout; plan Out line). Verified by the reviewers' probe and manual step 1.
Do: Give the framed child a view whatever its content. In `SourcePageArtifactRow.body`, wrap the `ArtifactCopyDetailsView(...)` call in `VStack(alignment: .leading, spacing: 0) { … }` before `.frame(minWidth: 0, maxWidth: Self.copyDetailsWidth(widthClass), alignment: .topLeading)`. Keep `ArtifactCopyDetailsView`'s doc comment true (it draws nothing; the row's container keeps the column). Add a one-line comment at the wrapper saying why it exists.
Check: scripts/check-pr.sh PASS.

**F-2 — T-1 covers the clamp** (from CON-3)
Files: `macos/ProvenenciaTests/PVFlowLayoutTests.swift`.
Failing test first: add `stillWiderThanTheRowIsPlacedAtTheRowWidth`: a `fit` that ignores the offer (ideal 300, re-measured 280) in a row 250 wide, with a second item. Expect the first frame's width 250 and the second item on the next line. Run it against a temporary removal of the clamp to see it red, then restore.
Do: test only.
Check: the test passes; scripts/check-pr.sh PASS.

**F-3 — Buttons go through `perform(_:)`** (from MNT-1)
Files: `macos/App/Features/Sources/SourcePageArtifactRow.swift`.
Do: The Set cover, pencil and trash buttons call `perform(.setCover)`, `perform(.edit)`, `perform(.delete)`. Compute `canUseAsThumbnail(art)` once in `body` and pass it to the menu and the actions.
Check: T-4 still passes; scripts/check-pr.sh PASS.

**F-4 — Tokens and a stale comment** (from MNT-3, MNT-4)
Files: `macos/App/DesignSystem/Components/Tag/PVTag.swift`, `macos/App/Features/Sources/SourceFormLoadArea.swift`.
Do: `PVTag`: `HStack(spacing: PVSpacing.space3)`, `.padding(.leading, PVSpacing.space4)`; keep 3, 9 and 7 as literals and say in the header comment they are the web kit's measures. `SourceFormLoadArea` doc comment: "for a card" → "for the Artifact sheet".
Check: scripts/check-pr.sh PASS.

**F-5 — Tooltips on truncated text** (from UX-2, UX-3, UX-4)
Files: `macos/App/Features/Sources/ArtifactCopyDetailsView.swift`, `macos/App/Features/Sources/SourcePageArtifactRow.swift`, `macos/App/Features/Conclusions/PersonSourcesSection.swift`.
Do: `.help(entry.field.label)` on the field-name `Text`; `.help(all.joined(separator: " · "))` on the graph-line `HStack`; `.help(label)` on the pill in `PersonSourcesSection.typeCell` (not inside `PVTag`).
Check: scripts/check-pr.sh PASS.

## Round 1 — fix
Head `b33f359a4`. Gate PASS (orchestrator run). F-1 – F-5 done. F-2's test was seen red with the clamp removed ("280.0 == 250.0"), then green. F-1 and F-3 share commit `c72d85c73`; UX-3's `.help` landed there too.
Minors read by the orchestrator: MNT-1 **Resolved** (buttons call `perform(_:)`, `canSetCover` computed once in `body`); MNT-3 **Resolved** (`space3`, `space4`, literals explained in the header comment); MNT-4 **Resolved** ("for the Artifact sheet").

## Round 2 — review
Reviewers: CON (CON-1 blocker), UX (UX-1 major). Scope `e771ec007...b33f359a4`. MNT and SEC+PRF not sent: no Fix-now blocker or major.

| ID | Lens | Severity | Title | Decision | Evidence / reason |
| --- | --- | --- | --- | --- | --- |
| CON-1, UX-1 | CON, UX | blocker | Copy-details column kept for an Artifact with no details | Resolved | both reviewers' hosted probe: text column 544pt with and without details once wrapped (888pt bare) |

CON: Clean. UX: Clean, no nits. No Fix plan 2.

## Result
Gate PASS (head `b33f359a4`, `origin/main` unmoved at `d18918601`) · CI pending at commit · open: none · follow-ups: none · to user: CON-2 (§ 10 manual check), MNT-2 (exclusive `title` / `verbatimTitle` vs D-13)
