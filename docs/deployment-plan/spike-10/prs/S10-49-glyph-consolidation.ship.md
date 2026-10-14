# Ship log — S10-49 Glyph consolidation

**Plan:** [`S10-49-glyph-consolidation.md`](S10-49-glyph-consolidation.md) · **Branch:** `s10-49-glyph-consolidation` · **PR:** #394
**State:** ready

## Before start

`main` @ `e150a76a`, the planning base. The branch already exists. Its first commit holds the plan, the review log, and the map entry, because the plan was not on `main` (review X-6). The implementer continues on this branch.

## Round 0 — implement
Head `2e3637a3`. PR #394 (draft). Gate PASS: the orchestrator ran `scripts/check-pr.sh` against `e150a76a`. Report deviations:
- AC-2.4's `rg` still matches the four bridge raw values in `enum PVMarkKey` (`PVMark.swift` 61–64). The plan's pattern was wrong: it matches the case definitions themselves. No other site matches. This is a plan defect, not a code defect.
- `ConclusionDetailPage.swift:232` built `.init(mark:label:)`, a caller §4 missed. It was moved to `glyph:` so the build passes.
- The `pbxproj` IDs for `PVGlyphTests.swift` are `…09E0` / `…09E1`. The implementer first tried `03C0`, which collided with `PVContextMenu.swift`.
- The S10-49 plan's Context and STEP-9 Files lines were repointed to `ideas/archive/` so `check-plan-docs` passes, the same as S10-19 did for its briefs.
- The manual check is not run. It is for the reviewer.

## Round 1 — review

All five lenses were dispatched against `e150a76a...2e3637a3`.

| ID | Lens | Severity | Title | Decision | Evidence / reason |
| --- | --- | --- | --- | --- | --- |
| MNT-1 | MNT | minor | `PVThumbnail` decides its glyph twice (`state` then `if case` re-match) | Follow-up | Verified at `PVThumbnail.swift` lines 57–65 and 83–91. D-4 kept the private `State` cases. No fix round touches the file. |
| CON-1 | CON | nit | AC-2.4's `rg` prints the 4 `PVMarkKey` raw-value lines | Logged | Plan defect, already in the Landed note and in the round-0 deviations |
| MNT-2 | MNT | nit | Unreachable `?? .fallback` on the omnibar person/event/place lead | Logged | Matches D-6 |
| MNT-3 | MNT | nit | Archived idea body still reads in the present tense | Logged | STEP-9 asked only for the title, status, links, and leftovers |
| CON-2 | CON | nit | Blank-line placement around the S10-49 map table | Logged | Renders correctly |
| UX | UX | — | No findings | — | — |
| SEC | SEC | — | No findings | — | — |
| PRF | PRF | — | No findings | — | — |

No Fix-now finding, so there is no Fix plan 1 and no round 2.

## Result
Gate PASS at `2e3637a3`. CI: see the PR. Open: none. Follow-ups: MNT-1 (`PVThumbnail` private `State` with associated values). To user: the manual check in plan §10.

## Fix round 1 — user-requested follow-ups (2026-10-14)

After the PR was marked ready, the user asked to fold the follow-ups into this PR ("implement those follow-ups on this same PR"). That covers MNT-1, the omnibar symbol drift, and the per-type switches the archived idea lists as leftovers. The user's request approves the omnibar change, which is visible: Source-type hits with no icon key show the tag symbol, and Metadata field hits show the list symbol, matching their sidebar rows. `FakeStore.syntheticPresentation` stays: it is the test double that stands in for the Go registry, so its table is the fake's data, not a copy to remove.

### Fix plan 1

**F-1 — `PVThumbnail` decides its glyph once** (from MNT-1)
Files: `macos/App/DesignSystem/Components/Thumbnail/PVThumbnail.swift`. Skill: `add-ui-component`.
Do: give the private `State` associated values (`case mark(PVMarkKey)`, `case symbol(PVSymbol)`, plus the existing `image`, `empty`, `loading`). `state` maps `content.glyph` once. `body` and `accessibilityLabel` draw and label from the matched value, with no `if case … = content.glyph` re-match and no `state == .markGlyph` guard. Sizes, inks, the decorative rule, and the label order stay exactly as they are (D-4). `isFilled` stays true for image, mark, and symbol.
Check: `rg -n "content\.glyph" macos/App/DesignSystem/Components/Thumbnail/PVThumbnail.swift` prints one line (the `state` mapping). Build succeeds. `ThumbnailCacheTests` pass.

**F-2 — Omnibar lead symbols match the sidebar** (archived-idea leftover; user approved)
Files: `macos/App/Features/Workspace/OmnibarHitPresentation.swift`, `macos/ProvenenciaTests/OmnibarResultsModelTests.swift`.
Do: in `leadThumbnail(for:)`, `"source_type"` with an empty `iconKey` → `Content(glyph: WorkspaceSection.sourceTypes.glyph)`, and `"metadata_field"` → `Content(glyph: WorkspaceSection.metadata.glyph)`. A `source_type` hit with an icon key keeps its own mark and label.
Failing test first: change the two expectations at `OmnibarResultsModelTests.swift:170–171` to `.symbol(.tag)` and `.symbol(.list)`, run them red, then fix.
Check: `OmnibarResultsModelTests` pass.

**F-3 — Properties reads palette order and bridge role from the presentation** (archived-idea leftover)
Files: `macos/App/Features/Catalog/SubjectTypeChrome.swift`, `macos/App/Features/Properties/PropertiesSnapshot.swift`, `macos/App/Features/Properties/PropertiesView.swift`, plus tests in `macos/ProvenenciaTests/PropertiesModelTests.swift`.
Do:
- Delete `PropertiesTypeChrome.sortIndex`. `PropertiesSnapshot.typesInPaletteOrder` sorts by `presentationsByKey[key]?.paletteSort ?? Int.max`, then by key for ties.
- Replace `showsBridgeLabel(typeKey:)` with `showsBridgeLabel(presentation: CatalogSubjectTypePresentation?) -> Bool`, true when `presentation?.role == "bridge"`.
- `ink(typeKey:presentation:)` mutes on that same check. Its `typeKey` fallback, used when no presentation is loaded (`EvidenceSubjectKindStyle.forKind`), stays.
- `PropertiesView.swift:179` passes `presentation`.

Failing test first: a `PropertiesModelTests` case that builds a snapshot whose `paletteSort` order differs from alphabetical, and expects `typesInPaletteOrder` in palette order. Another case expects `showsBridgeLabel(presentation:)` true only for role `bridge`, so `source` (role `reification`) is false.
Check: `rg -n "sortIndex|showsBridgeLabel\(typeKey" macos` prints nothing. `PropertiesModelTests` pass.

**F-4 — Composer group order from the typed kind** (archived-idea leftover)
Files: `macos/App/Features/CitationComposer/CitationComposerAuthoring.swift`.
Do: `typeRank(_:)` returns `EvidencePrimaryKind(rawValue: key).flatMap { EvidencePrimaryKind.allCases.firstIndex(of: $0) } ?? EvidencePrimaryKind.allCases.count`. The order stays Person, Event, Place, then everything else.
Check: build succeeds, and the composer tests pass.

**F-5 — Docs**
Files: `docs/ideas/archive/glyph-consolidation.md`, `docs/deployment-plan/spike-10/deployment-plan.md`, `docs/deployment-plan/spike-10/prs/S10-49-glyph-consolidation.md`, `docs/deployment-plan/spike-10/completed.md`.
Do:
- Archived idea: rewrite the two body sections in the past tense (MNT-3). Change the Leftovers section to say what F-1 to F-4 did. Keep one leftover: `FakeStore.syntheticPresentation` stays, as the Go registry's test double.
- Map: put one blank line between the S10-49 table and `#### S10-20`, and drop the extra blank line above `#### S10-49` (CON-2).
- Plan Landed note: add one line for this fix round.
- `completed.md` S10-49 paragraph: add one sentence. Omnibar Source-type and Metadata leads match their sidebar icons, and Properties palette order and bridge labels read the registry presentation.

Check: `python3 scripts/check-plan-docs.py docs/deployment-plan/spike-10/prs/S10-49-glyph-consolidation.md` reports no errors.

Gate: `scripts/check-pr.sh` PASS.

## Fix round 1 — result
Head `5b8dd376`. Gate PASS, run by the orchestrator. F-1 through F-5 are done. Deviation: the F-3 palette-order test was never seen failing on its own, because removing `sortIndex` stopped the target compiling first.

## Round 2 — review (scope `72f5ebed...5b8dd376`; CON, MNT, UX)

| ID | Lens | Severity | Title | Decision | Evidence / reason |
| --- | --- | --- | --- | --- | --- |
| CON-3 + UX-1 | CON, UX | blocker | The Properties strip and bound-type chips lose board order on real projects | Fix now → F-6 | Verified. `rg PaletteSort core` shows the registry sets it only for person, event, and place (`registry.go` lines 69, 80, 91). The other five default to 0 and tie, so the key tie-break puts Location first. `FakeStore` hides this because it gives them 3–7. |
| MNT (r2) | MNT | — | Clean. The literal `"bridge"` and the ink fallback with no presentation were noted below the bar. | Logged | `ink` with no presentation only arises for primary kinds (`EvidenceSubjectKindStyle.forKind`) |

### Fix plan 2

**F-6 — The registry carries the full palette order** (from CON-3, UX-1)
Files: `core/database/subjectvocab/registry.go`, `core/database/subjectvocab/subjectvocab_test.go`, `macos/ProvenenciaTests/PropertiesModelTests.swift`. Rule: `.cursor/rules/go-tests.mdc`.
Failing test first: `TestAllTypesPaletteSortFollowsBoardOrder` *(new, table-driven)* in `subjectvocab_test.go`. Sorting `AllTypes()` by `PaletteSort` gives `person, event, place, relationship, participation, location, place_relationship, source`, and no two types share a `PaletteSort`. Run it red: today six types tie at 0.
Do: in `registry.go`, set `PaletteSort: 3` on relationship, `4` on participation, `5` on location, `6` on place_relationship, and `7` on source. This matches the deleted `sortIndex` and `FakeStore.syntheticPresentation`. No migration: `PaletteSort` comes from the compiled seed, not a column. `PlaceableTypes()` keeps only roots, so it is unaffected. In `PropertiesModelTests`, add a case that ties two non-root types at the same `paletteSort` and expects the key tie-break, so the Swift sort's behaviour on ties is pinned.
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/database/subjectvocab/... ./api/ffi/...` passes. `PropertiesModelTests` pass. `scripts/check-pr.sh` PASS.

## Round 2 — fix
Head `d734c463`. Gate PASS, run by the orchestrator. F-6 is done. `TestAllTypesPaletteSortFollowsBoardOrder` was seen red first (six types tied at 0), then green. The registry sets `PaletteSort` 3–7 on the bridges and source. `PropertiesModelTests.typesInPaletteOrderBreaksTiesByKey` pins the Swift tie-break. Two review rounds are done, so there is no round 3.

## Result (after the user-requested follow-ups)
Gate PASS at `d734c463`. CI: see the PR. Open: none. Follow-ups: none from review. MNT-2 and the round-2 MNT notes are logged. To user: the manual check in plan §10, plus the Properties type strip order (Person, Event, Place, Relationship, Participation, Location, Place relationship, Source) and the omnibar Source-type and Metadata lead icons.
